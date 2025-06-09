package controller

import (
	"Food_recommendation/Basic/dao"
	"Food_recommendation/Basic/model"
	"Food_recommendation/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

func UserRegister(c *gin.Context) {
	var u model.User
	if err := c.BindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	if u.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user name is required"})
		return
	}
	if u.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}
	err := dao.CreateUser(c.Request.Context(), u)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": "user name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
	})
}
func UserLogin(c *gin.Context) {
	var u model.User
	if err := c.BindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	if u.Username == "" || u.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lack of data"})
		return
	}
	user, err := dao.UserLogin(c.Request.Context(), u)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login", "details": err.Error()})
		return
	}
	aToken, rToken, err := utils.GenToken(strconv.Itoa(int(user.ID)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "login successfully",
		"data": gin.H{
			"aToken": aToken,
			"rToken": rToken,
		},
	})
}

func SearchHandler(c *gin.Context) {
	keyword := c.Query("key")
	results, err := dao.UserSearch(c.Request.Context(), keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
	})
}
func DishHandler(c *gin.Context) {
	DID, _ := strconv.Atoi(c.Param("dishId"))
	SID, _ := strconv.Atoi(c.Param("storeId"))
	data, err := dao.GetADishes(c.Request.Context(), uint(SID), uint(DID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"name":     data.Name,
			"price":    data.Price,
			"decs":     data.Desc,
			"img":      data.ImageURL,
			"tags":     data.Tags,
			"rating":   data.AvgRating,
			"like_num": data.LikeNum,
		},
		"message": "success",
	})
}

//func Rating(c *gin.Context) {
//	uid, _ := strconv.Atoi(utils.ParseSet(c))
//	DID, _ := strconv.Atoi(c.Param("dishId"))
//	SID, _ := strconv.Atoi(c.Param("storeId"))
//}
