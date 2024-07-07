package user

import (
	"backend-go/middlewares/auth"
	"backend-go/models"
	"backend-go/utils"
	db2 "backend-go/utils/db"
	"backend-go/utils/kvdb"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func getUserInfo(c *gin.Context) {
	uid, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	kv := kvdb.GetKvDb()
	if err != nil {
		_ret := func() int64 {
			authHead := c.GetHeader("Authorization")
			if !strings.HasPrefix(authHead, "Bearer ") {
				return 0
			}
			token := strings.TrimPrefix(authHead, "Bearer ")
			_uid, err := (*kv).GetInt64("token-" + token)
			if err != nil {
				return 0
			}
			return _uid
		}()
		if _ret == 0 {
			c.JSON(http.StatusUnauthorized, models.DefaultUser)
			return
		}
		uid = _ret
	}
	var u models.User
	db := db2.GetDb()
	var cnt int64
	(*db).Model(models.User{}).Where("uid = ?", uid).Count(&cnt)
	if cnt == 0 {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("User Not Found."))
		return
	}
	(*db).Model(models.User{}).Limit(1).Where("uid = ?", uid).First(&u)
	c.JSON(http.StatusOK, u)
}

func getUserStat(c *gin.Context) {
	uid, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.DefaultUser)
		return
	}
	db := db2.GetDb()
	var cnt int64
	(*db).Model(models.User{}).Where("uid = ?", uid).Count(&cnt)
	if cnt == 0 {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("User Not Found."))
		return
	}
	var articles, collections int64
	(*db).Model(models.Article{}).Where("author = ?", uid).Count(&articles)
	(*db).Model(models.Collection{}).Where("author = ?", uid).Count(&collections)
	c.JSON(http.StatusOK, gin.H{
		"article_cnt":    articles,
		"collection_cnt": collections,
	})
}

func updateUserBasicInfo(c *gin.Context) {
	target, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid UID."))
		return
	}
	uid := c.GetInt64("uid")
	isAdmin := c.GetBool("role-admin")
	if target != uid && !isAdmin {
		c.JSON(http.StatusForbidden, utils.RequirePermission("user"))
		return
	}
	type req struct {
		Username string `json:"username" binding:"required"`
		Sign     string `json:"sign" binding:"required"`
	}
	var r req
	err = c.ShouldBindBodyWithJSON(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	var u models.User
	db := db2.GetDb()
	var cnt int64
	(*db).Model(models.User{}).Limit(1).Where("uid = ?", target).Count(&cnt)
	if cnt == 0 {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("User Not Found."))
		return
	}
	db.Model(models.User{}).Limit(1).Where("uid = ?", target).First(&u)
	db.Model(models.User{}).Where("uid = ?", target).Update("username", r.Username)
	db.Model(models.User{}).Where("uid = ?", target).Update("sign", r.Sign)
	c.JSON(http.StatusOK, gin.H{"uid": u.ID})
}

const ImageToken = "3|9U0P5wdkbllxWiK8crZRvQBxwSZa2akslupkcJB7"

func uploadAvatar(c *gin.Context) {
	target, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid UID."))
		return
	}
	uid := c.GetInt64("uid")
	isAdmin := c.GetBool("role-admin")
	if target != uid && !isAdmin {
		c.JSON(http.StatusForbidden, utils.RequirePermission("user"))
		return
	}
	avatar, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	if avatar.Size > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("File is too large."))
		return
	}
	var u models.User
	db := db2.GetDb()
	var cnt int64
	(*db).Model(models.User{}).Limit(1).Where("uid = ?", target).Count(&cnt)
	if cnt == 0 {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("User Not Found."))
		return
	}
	db.Model(models.User{}).Limit(1).Where("uid = ?", target).First(&u)
	var bufReader bytes.Buffer
	mpWriter := multipart.NewWriter(&bufReader)
	fw, err := mpWriter.CreateFormFile("file", avatar.Filename)
	if err != nil {
		fmt.Println("Create form file error: ", err)
		return
	}
	avatarFile, err := avatar.Open()
	if err != nil {
		fmt.Println("Create form file error: ", err)
		return
	}
	_, err = io.Copy(fw, avatarFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Internal Error")
		return
	}
	err = mpWriter.Close()
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Internal Error")
		return
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("POST", "https://image.allenyou.wang/api/v1/upload", &bufReader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Internal Error")
		return
	}
	req.Header.Set("Content-Type", mpWriter.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ImageToken))
	response, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Internal Error")
		return
	}
	if response.StatusCode != 200 {
		c.JSON(http.StatusInternalServerError, "Internal Error")
		return
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Internal Error")
		return
	}
	bodyObj := make(map[string]interface{})
	err = json.Unmarshal(body, &bodyObj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Internal Error")
		return
	}
	if !bodyObj["status"].(bool) {
		c.JSON(http.StatusInternalServerError, "Internal Error")
		return
	}
	dataObj := bodyObj["data"].(map[string]interface{})
	linksObj := dataObj["links"].(map[string]interface{})
	url := linksObj["url"].(string)
	db.Model(models.User{}).Where("uid = ?", target).Update("avatar_url", url)
	c.JSON(http.StatusOK, gin.H{"uid": u.ID})
}

func updatePassword(c *gin.Context) {
	target, err := strconv.ParseInt(c.Param("uid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid UID."))
		return
	}
	uid := c.GetInt64("uid")
	isAdmin := c.GetBool("role-admin")
	if target != uid && !isAdmin {
		c.JSON(http.StatusForbidden, utils.RequirePermission("user"))
		return
	}
	type req struct {
		PasswordHash string `json:"password_hash" binding:"required"`
	}
	var r req
	err = c.ShouldBindBodyWithJSON(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	var u models.User
	db := db2.GetDb()
	var cnt int64
	(*db).Model(models.User{}).Limit(1).Where("uid = ?", target).Count(&cnt)
	if cnt == 0 {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("User Not Found."))
		return
	}
	db.Model(models.User{}).Limit(1).Where("uid = ?", target).First(&u)
	newPassword := u.SetPassword(r.PasswordHash)
	db.Model(models.User{}).Where("uid = ?", target).Update("salted_password", newPassword)
	c.JSON(http.StatusOK, gin.H{"uid": u.ID})
}

func createUser(c *gin.Context) {
	type req struct {
		Username     string `json:"username" binding:"required"`
		Mail         string `json:"mail" binding:"required"`
		PasswordHash string `json:"password_hash" binding:"required"`
	}
	var r req
	err := c.ShouldBindBodyWithJSON(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	db := db2.GetDb()
	var cnt int64
	(*db).Model(models.User{}).Limit(1).Where("mail = ? or username = ?", r.Mail, r.Username).Count(&cnt)
	if cnt != 0 {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("User already exists."))
		return
	}
	u := models.CreateUser(r.Username, r.Mail, r.PasswordHash)
	db.Model(models.User{}).Create(&u)
	c.JSON(http.StatusOK, gin.H{"uid": u.ID})
}

func registerUser(c *gin.Context) {
	type req struct {
		Username     string `json:"username" binding:"required"`
		Mail         string `json:"mail" binding:"required"`
		PasswordHash string `json:"password_hash" binding:"required"`
		VerifyCode   string `json:"verify_code" binding:"required"`
	}
	var r req
	err := c.ShouldBindBodyWithJSON(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	kv := kvdb.GetKvDb()
	mailHash := fmt.Sprintf("%X", sha256.Sum256([]byte(r.Mail)))
	correctVerifyCode, err := (*kv).GetString("mail-" + mailHash)
	if err != nil || correctVerifyCode != r.VerifyCode {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid verify code."))
		return
	}
	err = (*kv).Delete("mail-" + mailHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildCommonResponse("Internal error."))
		return
	}
	db := db2.GetDb()
	var cnt int64
	(*db).Model(models.User{}).Limit(1).Where("mail = ? or username = ?", r.Mail, r.Username).Count(&cnt)
	if cnt != 0 {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("User already exists."))
		return
	}
	u := models.CreateUser(r.Username, r.Mail, r.PasswordHash)
	db.Model(models.User{}).Create(&u)
	c.JSON(http.StatusOK, gin.H{"uid": u.ID})
}

func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/users")
	g.GET("/:uid", getUserInfo)
	g.GET("/:uid/stat", getUserStat)
	g.GET("/", getUserInfo)
	g.POST("/", registerUser)
	pg := g.Group("/")
	pg.Use(auth.RequireLogin())
	pg.Use(auth.RequirePermission("user"))
	pg.POST("/:uid/info", updateUserBasicInfo)
	pg.POST("/:uid/avatar", uploadAvatar)
	pg.POST("/:uid/password", updatePassword)
	apg := g.Group("/")
	apg.Use(auth.RequireLogin())
	apg.Use(auth.RequirePermission("user.admin"))
	apg.POST("/admin/", createUser)
}
