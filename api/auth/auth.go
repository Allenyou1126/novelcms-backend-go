package auth

import (
	"backend-go/middlewares/auth"
	"backend-go/models"
	"backend-go/utils"
	db2 "backend-go/utils/db"
	"backend-go/utils/kvdb"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/gin-gonic/gin"
	rand2 "math/rand"
	"net/http"
	"slices"
	"strconv"
	"time"
)

func logout(c *gin.Context) {
	kv := kvdb.GetKvDb()
	err := (*kv).Delete("token-" + c.GetString("token"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildCommonResponse("Internal Error."))
		return
	}
	c.JSON(http.StatusOK, utils.BuildCommonResponse("Logged out."))
}

func login(c *gin.Context) {
	type req struct {
		Username  string `json:"username" binding:"required"`
		AuthCode  string `json:"auth-code" binding:"required"`
		Timestamp int64  `json:"timestamp" binding:"required"`
		Hash      string `json:"hash" binding:"required"`
	}
	var r req
	err := c.ShouldBindBodyWithJSON(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	if (int64)(time.Now().Sub(time.Unix(r.Timestamp, 0)).Abs().Seconds()) > 25 {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	a := strconv.FormatInt(r.Timestamp, 10) + r.AuthCode
	hsh := fmt.Sprintf("%X", sha256.Sum256([]byte(a)))
	if hsh != r.Hash {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	db := db2.GetDb()
	var user models.User
	var cnt int64
	db.Model(&models.User{}).Where("mail = ? or username = ?", r.Username, r.Username).Count(&cnt)
	if cnt == 0 {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Failed."))
		return
	}
	db.Model(&models.User{}).Where("mail = ? or username = ?", r.Username, r.Username).First(&user)
	authCode := fmt.Sprintf("%X", sha256.Sum256([]byte(user.Salt+r.AuthCode)))
	if user.SaltedPassword != authCode {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Failed."))
		return
	}
	if !slices.Contains(user.Permissions, "basic") {
		c.JSON(http.StatusForbidden, utils.BuildCommonResponse("Denied."))
		return
	}
	tokenBuf := make([]byte, 2048)
	_, err = rand.Reader.Read(tokenBuf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildCommonResponse("Internal Error."))
		return
	}
	token := fmt.Sprintf("%X", sha256.Sum256(tokenBuf))
	kv := kvdb.GetKvDb()
	err = (*kv).Set("token-"+token, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildCommonResponse("Internal Error."))
		return
	}
	err = (*kv).Expire("token-"+token, 30*24*60*60)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildCommonResponse("Internal Error."))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func verifyMail(c *gin.Context) {
	type req struct {
		Mail string `form:"mail" binding:"required"`
	}
	var r req
	err := c.ShouldBindQuery(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	mailHash := fmt.Sprintf("%X", sha256.Sum256([]byte(r.Mail)))
	kv := kvdb.GetKvDb()
	_, err = (*kv).Get("mail-" + mailHash)
	if err == nil {
		c.JSON(http.StatusForbidden, utils.BuildCommonResponse("Request rate limit reached!"))
		return
	}
	authCode := fmt.Sprintf("%06d", rand2.New(rand2.NewSource(time.Now().UnixNano())).Int31n(1000000))
	err = utils.SendMail(r.Mail, fmt.Sprintf("您在 Novel CMS 上的注册验证码是 %s", authCode), fmt.Sprintf("您在 Novel CMS 上的注册验证码是 <strong>%s</strong>", authCode))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildCommonResponse("Internal error."))
		return
	}
	err = (*kv).SetString("mail-"+mailHash, authCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildCommonResponse("Internal error."))
		return
	}
	err = (*kv).Expire("mail-"+mailHash, 300)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildCommonResponse("Internal error."))
		return
	}
	c.JSON(http.StatusOK, gin.H{"mail": r.Mail})
}

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/auth")
	g.POST("/user-token", login)
	g.GET("/verify-mail", verifyMail)
	requireLoginGroup := g.Group("/")
	requireLoginGroup.Use(auth.RequireLogin())
	requireLoginGroup.GET("/logout", logout)
}
