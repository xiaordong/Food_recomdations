package utils

import "github.com/gin-gonic/gin"

func AuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		aToken := c.GetHeader("Authorization")
		if aToken == "" {
			c.JSON(401, "Unauthorized")
			c.Abort()
			return
		}
		if claims, err := ParasToken(aToken); err == nil {
			c.Set("id", claims.ID)
			c.Next()
		} else {
			c.JSON(401, "failed authorize")
			c.Abort()
			return
		}
	}
}
