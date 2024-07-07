package api

import (
	"backend-go/api/article"
	"backend-go/api/auth"
	"backend-go/api/collection"
	"backend-go/api/user"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	auth.RegisterRoutes(r)
	article.RegisterRoutes(r)
	collection.RegisterRoutes(r)
	user.RegisterRoutes(r)
}
