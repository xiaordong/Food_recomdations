package controller

import (
	"Food_recommendation/Basic/dao"
	"Food_recommendation/Basic/model"
	"Food_recommendation/utils"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"net/http"
	"strconv"
	"strings"
)

func NewStore(c *gin.Context) {
	var s model.Store
	if err := c.ShouldBind(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	if s.Name == "" || s.Description == "" || s.Address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": "lack key information"})
		return
	}
	ID, _ := strconv.Atoi(utils.ParseSet(c))
	s.MerchantID = uint(ID)
	if err := dao.CreateStore(c.Request.Context(), s); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": "Store name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Store", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Store created successfully",
	})
}
func GetStores(c *gin.Context) {
	ID, _ := strconv.Atoi(utils.ParseSet(c))
	data, err := dao.MyStore(c.Request.Context(), uint(ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Store", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully get Store",
		"data":    data,
	})
}
func AStore(c *gin.Context) {
	SID, err := strconv.Atoi(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store ID"})
		return
	}

	store, err := dao.SearchStore(c.Request.Context(), uint(SID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get store",
			"details": err.Error(),
		})
		return
	}

	// 构建精简响应
	response := gin.H{
		"id":          store.ID,
		"name":        store.Name,
		"description": store.Description,
		"active":      store.Active,
		"avgRating":   store.AvgRating,
		"address":     store.Address,
		"dishes":      formatDishes(store.Dishes),
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully get store",
		"data":    response,
	})

}

// 格式化菜品数据，过滤冗余字段
func formatDishes(dishes []model.Dishes) []gin.H {
	result := make([]gin.H, 0, len(dishes))

	for _, dish := range dishes {
		// 跳过嵌套的空store对象
		dish.Store = model.Store{}

		// 格式化标签（仅保留名称）
		tags := make([]string, 0, len(dish.Tags))
		for _, tag := range dish.Tags {
			tags = append(tags, tag.Name)
		}

		// 构建精简的菜品对象
		formattedDish := gin.H{
			"id":        dish.ID,
			"name":      dish.Name,
			"price":     dish.Price.String(), // 保留原始精度
			"desc":      dish.Desc,
			"imageUrl":  dish.ImageURL,
			"available": dish.Available,
			"avgRating": dish.AvgRating,
			"likeNum":   dish.LikeNum,
			"tags":      tags,
		}

		result = append(result, formattedDish)
	}

	return result
}
func UpdateStore(c *gin.Context) {
	MID, _ := strconv.Atoi(utils.ParseSet(c))
	SID, _ := strconv.Atoi(c.Param("storeId"))
	var s model.Store
	if err := c.ShouldBind(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	s.MerchantID = uint(MID)
	s.ID = uint(SID)
	if err := dao.UpdateStore(c.Request.Context(), s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update Store", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully update Store",
	})
}
func DeleteStore(c *gin.Context) {
	MID, _ := strconv.Atoi(utils.ParseSet(c))
	SID, _ := strconv.Atoi(c.Param("storeId"))
	if err := dao.DeleteStore(c.Request.Context(), uint(SID), (uint)(MID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete Store", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully delete Store",
	})
}
func NewDishes(c *gin.Context) {
	// 解析storeId参数
	SID, err := strconv.Atoi(c.Param("storeId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store ID", "details": err.Error()})
		return
	}

	// 定义DTO接收前端数据
	type RequestBody struct {
		Name      string   `json:"name" binding:"required"`
		Price     string   `json:"price" binding:"required"`
		Desc      string   `json:"desc"`
		ImageURL  string   `json:"imageUrl"`
		Tags      []string `json:"tags"`
		Available bool     `json:"available"`
	}

	var req RequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// 转换价格为decimal.Decimal类型
	price, err := decimal.NewFromString(req.Price)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price format", "details": err.Error()})
		return
	}

	// 构建完整的Dishes对象
	dishes := model.Dishes{
		StoreID:   uint(SID),
		Name:      req.Name,
		Price:     price,
		Desc:      req.Desc,
		ImageURL:  req.ImageURL,
		Available: req.Available,
	}

	// 处理Tags - 将字符串数组转换为Tag对象数组
	if len(req.Tags) > 0 {
		tags := make([]model.Tag, len(req.Tags))
		for i, tagName := range req.Tags {
			tags[i] = model.Tag{
				Name: tagName,
			}
		}
		dishes.Tags = tags
	}

	// 调用DAO层函数创建菜品
	if err := dao.CreateDishes(c.Request.Context(), dishes); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": "Dishes name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Dishes", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully created Dishes",
	})
}
func GetDishes(c *gin.Context) {
	SID, _ := strconv.Atoi(c.Param("storeId"))
	MID, _ := strconv.Atoi(utils.ParseSet(c))
	data, err := dao.GetDishes(c.Request.Context(), uint(SID), uint(MID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Dishes", "details": err.Error()})
		return
	}
	var response []struct {
		ID        uint            `json:"id"`
		StoreID   uint            `json:"storeId"`
		Name      string          `json:"name"`
		Price     decimal.Decimal `json:"price"`
		Desc      string          `json:"desc"`
		ImageURL  string          `json:"imageUrl"`
		LikeNum   uint            `json:"likeNum"`
		Available bool            `json:"available"`
	}
	for _, dish := range data {
		response = append(response, struct {
			ID        uint            `json:"id"`
			StoreID   uint            `json:"storeId"`
			Name      string          `json:"name"`
			Price     decimal.Decimal `json:"price"`
			Desc      string          `json:"desc"`
			ImageURL  string          `json:"imageUrl"`
			LikeNum   uint            `json:"likeNum"`
			Available bool            `json:"available"`
		}{
			ID:        dish.ID,
			StoreID:   dish.StoreID,
			Name:      dish.Name,
			Price:     dish.Price,
			Desc:      dish.Desc,
			ImageURL:  dish.ImageURL,
			LikeNum:   dish.LikeNum,
			Available: dish.Available,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully get Dishes",
		"data":    response,
	})
}
