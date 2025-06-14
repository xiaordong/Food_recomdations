package controller

import (
	"Food_recommendation/Basic/dao"
	"Food_recommendation/Basic/model"
	"Food_recommendation/utils"
	"fmt"
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
	user.Password = ""
	user.Phone = user.Phone[:3] + "****" + user.Phone[7:]
	c.JSON(http.StatusOK, gin.H{
		"message": "login successfully",
		"data": gin.H{
			"aToken": aToken,
			"rToken": rToken,
		},
		"info": user,
	})
}

func SearchHandler(c *gin.Context) {
	keyword := c.Query("key")
	token := c.GetHeader("Authorization")
	uid := 0
	if token != "" {
		claim, err := utils.ParasToken(token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login", "details": err.Error()})
			return
		}
		uid, _ = strconv.Atoi(claim.ID)
	}
	results, err := dao.UserSearch(c.Request.Context(), keyword, uint(uid))
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

//	func Rating(c *gin.Context) {
//		uid, _ := strconv.Atoi(utils.ParseSet(c))
//		DID, _ := strconv.Atoi(c.Param("dishId"))
//		SID, _ := strconv.Atoi(c.Param("storeId"))
//	}
func LikeDishHandler(c *gin.Context) {
	var req struct {
		DishID uint `json:"dishId"`
		IsLike bool `json:"isLike"` // true=点赞，false=取消点赞
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}
	userID, _ := strconv.Atoi(utils.ParseSet(c))
	if err := dao.LikeDish(c.Request.Context(), uint(userID), req.DishID, req.IsLike); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Like failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Like operation succeeded"})
}
func RateDishHandler(c *gin.Context) {
	var req struct {
		DishID uint   `json:"dishId" binding:"required"`
		Score  uint   `json:"score" binding:"required,min=1,max=5"` // 评分范围1-5
		Commit string `json:"commit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	fmt.Println(req)
	userID, _ := strconv.Atoi(utils.ParseSet(c))

	if err := dao.RateDish(c.Request.Context(), uint(userID), req.DishID, req.Score, req.Commit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rate submitted successfully", "avgRating": req.Score})
}
func GetHistory(c *gin.Context) {
	userID, _ := strconv.Atoi(utils.ParseSet(c))
	res, err := dao.GetUserHistory(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user history", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully get user history",
		"data":    res,
	})
}
func GetSearchKey(c *gin.Context) {
	uid, _ := strconv.Atoi(utils.ParseSet(c))
	res, err := dao.AllSearch(c.Request.Context(), uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully",
		"data":    res,
	})
}
func UserLike(c *gin.Context) {
	uid, _ := strconv.Atoi(utils.ParseSet(c))
	res, err := dao.AllLike(c.Request.Context(), uint(uid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully",
		"data":    res,
	})

}
