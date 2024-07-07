package auth

import (
	"backend-go/models"
	"backend-go/utils"
	db2 "backend-go/utils/db"
	"backend-go/utils/kvdb"
	"github.com/gin-gonic/gin"
	"net/http"
	"slices"
	"strings"
)

func RequireLogin() gin.HandlerFunc {
	kv := kvdb.GetKvDb()
	return func(c *gin.Context) {
		authHead := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHead, "Bearer ") {
			c.JSON(http.StatusUnauthorized, utils.RequireToken())
			c.Abort()
			return
		}
		token := strings.TrimPrefix(authHead, "Bearer ")
		uid, err := (*kv).GetInt64("token-" + token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.InvalidToken())
			c.Abort()
			return
		}
		c.Set("uid", uid)
		c.Set("token", token)
		c.Next()
	}
}

func RequirePermission(requiredPermission string) gin.HandlerFunc {
	db := db2.GetDb()
	return func(c *gin.Context) {
		uid := c.GetInt64("uid")
		// TODO: Get permission list of user with uid here
		var u models.User
		(*db).Model(models.User{}).Where("uid = ?", uid).First(&u)
		if !func() bool {
			if !slices.Contains(u.Permissions, requiredPermission) {
				return false
			}
			c.Set("permission-"+requiredPermission, true)
			if slices.Contains(u.Permissions, requiredPermission+".admin") || strings.HasSuffix(requiredPermission, ".admin") {
				c.Set("role-admin", true)
			}
			return true
		}() {
			c.JSON(http.StatusForbidden, utils.RequirePermission(requiredPermission))
			c.Abort()
			return
		}
		c.Next()
	}
}
