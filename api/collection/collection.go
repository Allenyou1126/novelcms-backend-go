package collection

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

func getCollectionList(c *gin.Context) {
	type req struct {
		PageSize int   `json:"page-size"`
		Page     int   `json:"page"`
		Author   int64 `json:"author" form:"author"`
	}
	var r req
	err := c.ShouldBind(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	if r.PageSize == 0 {
		r.PageSize = 10
	}
	if r.Page == 0 {
		r.Page = 1
	}
	if r.Page < 0 || r.PageSize < 0 || r.PageSize > 50 {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	db := db2.GetDb().Model(models.Collection{})
	if r.Author != 0 {
		db = db.Where("author = ?", r.Author)
	}
	var cnt int64 = 0
	(*db).Count(&cnt)
	ret := make([]models.Collection, 1)
	res := (*db).Preload("Author").Order("updated_at desc").Limit(r.PageSize).Offset(r.PageSize * (r.Page - 1)).Find(&ret)
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("Collections Not Found."))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"count":       res.RowsAffected,
		"total":       math.Ceil(float64(cnt) / float64(r.PageSize)),
		"collections": ret,
	})
}

func getCollection(c *gin.Context) {
	cid, err := strconv.ParseInt(c.Param("cid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("Collection Not Found."))
		return
	}
	var a models.Collection
	db := db2.GetDb()
	var cnt int64
	(*db).Model(models.Collection{}).Limit(1).Where("cid = ?", cid).Count(&cnt)
	if cnt == 0 {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("Collection Not Found."))
		return
	}
	(*db).Model(models.Collection{}).Preload("Author").Limit(1).Where("cid = ?", cid).First(&a)
	c.JSON(http.StatusOK, a)
}

func getCollectionArticles(c *gin.Context) {
	cid, err := strconv.ParseInt(c.Param("cid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("Collection Not Found."))
		return
	}
	type req struct {
		PageSize int `json:"page-size" form:"page-size"`
		Page     int `json:"page" form:"page"`
	}
	var r req
	err = c.ShouldBindQuery(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	if r.PageSize == 0 {
		r.PageSize = 10
	}
	if r.Page == 0 {
		r.Page = 1
	}
	if r.Page < 0 || r.PageSize < 0 || r.PageSize > 50 {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	db := db2.GetDb()
	var cnt int64 = 0
	(*db).Model(models.Article{}).Where("collection_id = ?", cid).Count(&cnt)
	ret := make([]models.Article, 1)
	res := (*db).Model(models.Article{}).Preload("Author").Order("updated_at desc").Where("collection_id = ?", cid).Offset(r.PageSize * (r.Page - 1)).Limit(r.PageSize).Find(&ret)
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

func createCollection(c *gin.Context) {
	type req struct {
		Title       string `binding:"required"`
		Description string
	}
	var r req
	err := c.ShouldBindBodyWithJSON(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	uid := c.GetInt64("uid")
	a := models.CreateCollection(uid, r.Title)
	if r.Description != "" {
		a.Description = &r.Description
	} else {
		a.Description = nil
	}
	db := db2.GetDb()
	db.Model(models.Collection{}).Create(&a)
	c.JSON(http.StatusOK, gin.H{"cid": a.ID})
}

func updateCollection(c *gin.Context) {
	cid, err := strconv.ParseInt(c.Param("cid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("Collection Not Found."))
		return
	}
	type req struct {
		Title       string `binding:"required"`
		Description string
	}
	var r req
	err = c.ShouldBindBodyWithJSON(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	uid := c.GetInt64("uid")
	var a models.Collection
	db := db2.GetDb()
	var cnt int64
	(*db).Model(models.Collection{}).Limit(1).Where("cid = ?", cid).Count(&cnt)
	if cnt == 0 {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("Collection Not Found."))
		return
	}
	db.Model(models.Collection{}).Limit(1).Where("cid = ?", cid).First(&a)
	if a.AuthorID != uid && !c.GetBool("role-admin") {
		c.JSON(http.StatusForbidden, utils.RequirePermission("collection.admin"))
		return
	}
	a.Title = r.Title
	if r.Description != "" {
		a.Description = &r.Description
	} else {
		a.Description = nil
	}
	db.Model(models.Collection{}).Save(&a)
	c.JSON(http.StatusOK, gin.H{"cid": a.ID})
}

func deleteCollection(c *gin.Context) {
	cid, err := strconv.ParseInt(c.Param("cid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("Collection Not Found."))
		return
	}
	uid := c.GetInt64("uid")
	var a models.Collection
	db := db2.GetDb()
	var cnt int64
	(*db).Model(models.Collection{}).Limit(1).Where("cid = ?", cid).Count(&cnt)
	if cnt == 0 {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("Collection Not Found."))
		return
	}
	db.Model(models.Collection{}).Limit(1).Where("cid = ?", cid).First(&a)
	if a.AuthorID != uid && !c.GetBool("role-admin") {
		c.JSON(http.StatusForbidden, utils.RequirePermission("collection.admin"))
		return
	}
	db.Model(models.Collection{}).Delete(&a)
	c.JSON(http.StatusOK, gin.H{"cid": a.ID})
}

func insertArticleIntoCollection(c *gin.Context) {
	cid, err := strconv.ParseInt(c.Param("cid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("Collection Not Found."))
		return
	}
	type req struct {
		TargetArticles []int64 `json:"target_articles" binding:"required"`
	}
	var r req
	err = c.ShouldBindBodyWithJSON(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	uid := c.GetInt64("uid")
	var a models.Collection
	db := db2.GetDb()
	var cnt int64
	(*db).Model(models.Collection{}).Limit(1).Where("cid = ?", cid).Count(&cnt)
	if cnt == 0 {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("Collection Not Found."))
		return
	}
	db.Model(models.Collection{}).Limit(1).Where("cid = ?", cid).First(&a)
	if a.AuthorID != uid && !c.GetBool("role-admin") {
		c.JSON(http.StatusForbidden, utils.RequirePermission("collection.admin"))
		return
	}
	var as []models.Article
	errCnt := len(r.TargetArticles)
	if c.GetBool("role-admin") {
		db.Model(models.Article{}).Where("aid IN ? AND (collection_id = 0 OR collection_id = NULL)", r.TargetArticles).Find(&as)
	} else {
		db.Model(models.Article{}).Where("aid IN ? AND (collection_id = 0 OR collection_id = NULL) AND author = ?", r.TargetArticles, uid).Find(&as)
	}
	err = db.Model(&a).Association("Articles").Append(as)
	if err == nil {
		errCnt -= len(as)
	}
	c.JSON(http.StatusOK, gin.H{"cid": a.ID, "error_cnt": errCnt})
}

func deleteArticleFromCollection(c *gin.Context) {
	cid, err := strconv.ParseInt(c.Param("cid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("Collection Not Found."))
		return
	}
	type req struct {
		TargetArticles []int64 `json:"target_articles" binding:"required"`
	}
	var r req
	err = c.ShouldBindBodyWithJSON(&r)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildCommonResponse("Invalid request."))
		return
	}
	uid := c.GetInt64("uid")
	var a models.Collection
	db := db2.GetDb()
	var cnt int64
	(*db).Model(models.Collection{}).Limit(1).Where("cid = ?", cid).Count(&cnt)
	if cnt == 0 {
		c.JSON(http.StatusNotFound, utils.BuildCommonResponse("Collection Not Found."))
		return
	}
	db.Model(models.Collection{}).Limit(1).Where("cid = ?", cid).First(&a)
	if a.AuthorID != uid && !c.GetBool("role-admin") {
		c.JSON(http.StatusForbidden, utils.RequirePermission("collection.admin"))
		return
	}
	var as []models.Article
	errCnt := len(r.TargetArticles)
	if c.GetBool("role-admin") {
		db.Model(models.Article{}).Where("aid IN ? AND collection_id = ?", r.TargetArticles, cid).Find(&as)
	} else {
		db.Model(models.Article{}).Where("aid IN ? AND collection_id = ? AND author = ?", r.TargetArticles, cid, uid).Find(&as)
	}
	err = db.Model(&a).Association("Articles").Delete(as)
	if err == nil {
		errCnt -= len(as)
	}
	c.JSON(http.StatusOK, gin.H{"cid": a.ID, "error_cnt": errCnt})
}

func RegisterRoutes(r *gin.Engine) {
	r.GET("/collections", getCollectionList)
	g := r.Group("/collections")
	g.GET("/", getCollectionList)
	g.GET("/:cid", getCollection)
	g.GET("/:cid/articles", getCollectionArticles)
	pg := g.Group("/")
	pg.Use(auth.RequireLogin())
	pg.Use(auth.RequirePermission("collection"))
	pg.POST("/", createCollection)
	pg.PUT("/:cid", updateCollection)
	pg.DELETE("/:cid", deleteCollection)
	pg.PUT("/:cid/articles", insertArticleIntoCollection)
	pg.DELETE("/:cid/articles", deleteArticleFromCollection)
}
