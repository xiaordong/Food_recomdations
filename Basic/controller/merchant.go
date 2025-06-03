package controller

import (
	"Food_recommendation/Basic/dao"
	"Food_recommendation/Basic/model"
	"Food_recommendation/utils"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func MerchantRegister(c *gin.Context) {
	var m model.Merchant
	if err := c.BindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	log.Printf("Parsed merchant: %+v", m)

	if m.MerchantName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Merchant name is required"})
		return
	}
	if m.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	// 调用创建商户函数
	err := dao.MerchantCreate(c.Request.Context(), m)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": "Merchant name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create merchant", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Merchant registered successfully",
		"merchant": gin.H{
			"merchantName": m.MerchantName,
			"store":        nil,
		},
	})
}
func MerchantLogin(c *gin.Context) {
	var m model.Merchant
	if err := c.BindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	if m.MerchantName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Merchant name is required"})
		return
	}
	if m.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}
	if err := dao.CheckLogin(c, m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Merchant login failed", "details": err.Error()})
		return
	}
	aToken, rToken, err := utils.GenToken(strconv.Itoa(int(m.ID)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to generate token", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Merchant login successfully",
		"data": gin.H{
			"aToken": aToken,
			"rToken": rToken,
		},
	})
}

func GetMerchant(c *gin.Context) {
	m, err := dao.GetProfile(c.Request.Context(), utils.ParseSet(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Merchant not found", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"merchant": gin.H{
			"merchantID":   m.ID,
			"merchantName": m.MerchantName,
			"phone":        m.Phone,
			"store":        m.Stores,
		},
	})
}

func UpdateMerchant(c *gin.Context) {
	var m model.Merchant
	if err := c.BindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}
	ID, _ := strconv.Atoi(utils.ParseSet(c))
	m.ID = uint(ID)
	err := dao.UpdateProfile(c.Request.Context(), m)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update merchant", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Merchant updated successfully",
	})

}
