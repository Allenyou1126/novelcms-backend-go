package article

import (
	"backend-go/middlewares/auth"
	"backend-go/models"
	"backend-go/utils"
	db2 "backend-go/utils/db"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"strconv"
)

func getArticleList(c *gin.Context) {
	type req struct {
		PageSize int   `json:"page-size" form:"page-size"`
		Page     int   `json:"page" form:"page"`
		Author   int64 `json:"author" form:"author"`
	}
	var r req
	err := c.ShouldBindQuery(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	if r.Page < 0 || r.PageSize < 0 || r.PageSize > 50 {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	db := db2.GetDb().Model(models.Article{})
	if r.Author != 0 {
		db = db.Where("author = ?", r.Author)
	}
	var cnt int64 = 0
	(*db).Count(&cnt)
	if r.PageSize == 0 {
		r.PageSize = int(cnt)
	}
	if r.Page == 0 {
		r.Page = 1
	}
	ret := make([]models.Article, 1)
	res := (*db).Preload("Author").Order("updated_at desc").Offset(r.PageSize * (r.Page - 1)).Limit(r.PageSize).Find(&ret)
	if res.RowsAffected == 0 || int(cnt) <= r.PageSize*(r.Page-1) {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("Articles Not Found."))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"count":    cnt,
		"total":    math.Ceil(float64(cnt) / float64(r.PageSize)),
		"articles": ret,
	})
}

func getArticle(c *gin.Context) {
	aid, err := strconv.ParseInt(c.Param("aid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("Article Not Found."))
		return
	}
	var a models.Article
	db := db2.GetDb()
	var cnt int64
	(*db).Model(models.Article{}).Where("aid = ?", aid).Count(&cnt)
	if cnt == 0 {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("Article Not Found."))
		return
	}
	(*db).Model(models.Article{}).Preload("Author").Limit(1).Where("aid = ?", aid).First(&a)
	c.JSON(http.StatusOK, a)
}

func createArticle(c *gin.Context) {
	type req struct {
		Title   string `binding:"required"`
		Content string `binding:"required"`
	}
	var r req
	err := c.ShouldBindBodyWithJSON(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	uid := c.GetInt64("uid")
	a := models.CreateArticle(uid, r.Title, r.Content)
	db := db2.GetDb()
	db.Model(models.Article{}).Create(&a)
	c.JSON(http.StatusOK, gin.H{"aid": a.ID})
}

func updateArticle(c *gin.Context) {
	aid, err := strconv.ParseInt(c.Param("aid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("Article Not Found."))
		return
	}
	type req struct {
		Title   string `binding:"required"`
		Content string `binding:"required"`
	}
	var r req
	err = c.ShouldBindBodyWithJSON(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	uid := c.GetInt64("uid")
	var a models.Article
	db := db2.GetDb()
	db.Model(models.Article{}).Limit(1).Where("aid = ?", aid).First(&a)
	if a.AuthorID != uid && !c.GetBool("role-admin") {
		c.JSON(http.StatusForbidden, utils.RequirePermission("article.admin"))
		return
	}
	a.Title = r.Title
	a.Content = r.Content
	db.Save(&a)
	c.JSON(http.StatusOK, gin.H{"aid": a.ID})
}

func deleteArticle(c *gin.Context) {
	aid, err := strconv.ParseInt(c.Param("aid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("Article Not Found."))
		return
	}
	uid := c.GetInt64("uid")
	var a models.Article
	db := db2.GetDb()
	db.Model(models.Article{}).Limit(1).Where("aid = ?", aid).First(&a)
	if a.AuthorID != uid && !c.GetBool("role-admin") {
		c.JSON(http.StatusForbidden, utils.RequirePermission("article.admin"))
		return
	}
	db.Model(models.Article{}).Delete(&a)
	c.JSON(http.StatusOK, gin.H{"aid": a.ID})
}

func RegisterRoutes(r *gin.Engine) {
	r.GET("/articles", getArticleList)
	g := r.Group("/articles")
	g.GET("/", getArticleList)
	g.GET("/:aid", getArticle)
	pg := g.Group("/")
	pg.Use(auth.RequireLogin())
	pg.Use(auth.RequirePermission("article"))
	pg.POST("/", createArticle)
	pg.PUT("/:aid", updateArticle)
	pg.DELETE("/:aid", deleteArticle)
}
