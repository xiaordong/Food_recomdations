package router

import (
	"Food_recommendation/Basic/controller"
	"Food_recommendation/utils"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	router := gin.Default()
	merchant := router.Group("/api/merchant")
	merchant.POST("/register", controller.MerchantRegister)
	merchant.POST("/login", controller.MerchantLogin)
	authMerchant := merchant.Group("/")
	authMerchant.Use(utils.AuthMiddleware())
	{
		//商家信息管理
		authMerchant.GET("/profile", controller.GetMerchant)
		authMerchant.PUT("/profile", controller.UpdateMerchant)
		//店铺管理
		authMerchant.POST("/stores")
		authMerchant.GET("/stores")
		//店铺详情管理
		store := authMerchant.Group("/store/:storeId")
		{
			store.GET("/")
			store.PUT("/")
			store.DELETE("/")
			//菜品管理
			store.POST("/dishes")
			store.GET("/dishes")
			//菜品详情管理
			dish := store.Group("/dishes/:dishId")
			{
				dish.GET("/")
				dish.PUT("/")
				dish.DELETE("/")
				dish.PATCH("/description")
				//菜品标签管理
				dish.POST("/tags")
				dish.GET("/tags")
				dish.PUT("/tags")
				dish.DELETE("/tags/:tagId")
			}
		}
	}
	return router
}
