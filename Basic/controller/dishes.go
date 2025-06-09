package controller

import (
	"Food_recommendation/Basic/dao"
	"Food_recommendation/Basic/model"
	"Food_recommendation/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"net/http"
	"strconv"
)

func GetADishes(c *gin.Context) {
	// 解析参数
	SID, _ := strconv.Atoi(c.Param("storeId"))
	DID, _ := strconv.Atoi(c.Param("dishId"))

	// 调用DAO层获取包含标签的菜品数据
	data, err := dao.GetADishes(c.Request.Context(), uint(SID), uint(DID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dish", "details": err.Error()})
		return
	}

	// 构建响应结构体，包含标签字段
	response := struct {
		ID        uint            `json:"id"`
		StoreID   uint            `json:"storeId"`
		Name      string          `json:"name"`
		Price     decimal.Decimal `json:"price"`
		Desc      string          `json:"desc"`
		ImageURL  string          `json:"imageUrl"`
		Available bool            `json:"available"`
		Rating    float64         `json:"rating"`
		LikeNum   uint            `json:"likeNum"`
		Tags      []string        `json:"tags"` // 新增标签字段
	}{
		ID:        data.ID,
		StoreID:   data.StoreID,
		Name:      data.Name,
		Price:     data.Price,
		Desc:      data.Desc,
		ImageURL:  data.ImageURL,
		Available: data.Available,
		Rating:    data.AvgRating,
		LikeNum:   data.LikeNum,
		Tags: func() []string {
			var tags []string
			for _, tag := range data.Tags {
				tags = append(tags, tag.Name)
			}
			return tags
		}(),
	}

	// 返回响应
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully get dish",
		"data":    response,
	})
}
func UpdateADishes(c *gin.Context) {
	SID, _ := strconv.Atoi(c.Param("storeId"))
	DID, _ := strconv.Atoi(c.Param("dishId"))
	var d model.Dishes
	if err := c.BindJSON(&d); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	d.ID = uint(DID)
	d.StoreID = uint(SID)
	if err := dao.UpdateDishes(c.Request.Context(), d); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Dishes", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully update Dishes",
	})
}
func DeleteADishes(c *gin.Context) {
	SID, _ := strconv.Atoi(c.Param("storeId"))
	DID, _ := strconv.Atoi(c.Param("dishId"))
	MID, _ := strconv.Atoi(utils.ParseSet(c))
	if !dao.Check(c.Request.Context(), uint(SID), uint(MID)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unauthorized"})
		return
	}
	if err := dao.DeleteDishes(c.Request.Context(), uint(SID), uint(DID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete Dishes", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully delete Dishes",
	})
}
func AddTags(c *gin.Context) {
	SID, _ := strconv.Atoi(c.Param("storeId"))
	DID, _ := strconv.Atoi(c.Param("dishId"))
	if !dao.Check(c.Request.Context(), uint(SID), uint(DID)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unauthorized"})
		return
	}
	var req struct {
		Tags []string `json:"tags" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json data"})
		return
	}
	fmt.Println(req.Tags)
	if err := dao.AddTags(c.Request.Context(), req.Tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "add tags failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "tags added successfully"})
}
func GetTags(c *gin.Context) {
	data, err := dao.GetTags(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tags", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully get tags",
		"data":    data,
	})

}
func ChooseTags(c *gin.Context) {
	SID, _ := strconv.Atoi(c.Param("storeId"))
	DID, _ := strconv.Atoi(c.Param("dishId"))
	MID, _ := strconv.Atoi(utils.ParseSet(c))
	if SID <= 0 || DID <= 0 || MID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameters"})
		return
	}
	// 解析请求体
	var req struct {
		Tags []string `json:"tags" binding:"required,max=3"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}
	// 检查菜品是否存在且属于当前商户
	if !dao.Check(c.Request.Context(), uint(SID), uint(MID)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unauthorized"})
		return
	}
	if err := dao.ChooseTag(c.Request.Context(), uint(DID), req.Tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set tags", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "tags set successfully",
		"dishId":  DID,
		"tags":    req.Tags,
	})
}
