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
	MID, _ := strconv.Atoi(utils.ParseSet(c))
	SID, _ := strconv.Atoi(c.Param("storeId"))
	data, err := dao.SearchStore(c.Request.Context(), uint(SID), (uint)(MID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Store", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully get Store",
		"data":    data,
	})
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
