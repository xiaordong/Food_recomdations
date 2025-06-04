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

func NewStore(c *gin.Context) {
	var s model.Store
	if err := c.ShouldBind(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	fmt.Println(s)
	if s.Name == "" || s.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": "lack of Name or Description"})
		return
	}
	ID, _ := strconv.Atoi(utils.ParseSet(c))
	s.MerchantID = uint(ID)
	fmt.Println(s)
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
		"data":    s,
	})
}
