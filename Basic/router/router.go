package router

import (
	"Food_recommendation/Basic/controller"
	"Food_recommendation/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func InitRouter() *gin.Engine {
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}                                                 // 允许前端域名
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}           // 允许的HTTP方法
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"} // 允许的请求头
	config.AllowCredentials = true                                                      // 允许携带凭证（如cookie）
	config.MaxAge = 12 * time.Hour
	router.Use(cors.New(config))
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
		authMerchant.POST("/stores", controller.NewStore)
		authMerchant.GET("/stores", controller.GetStores)
		//店铺详情管理
		store := authMerchant.Group("/store/:storeId")
		{
			store.GET("/", controller.AStore)
			store.PUT("/", controller.UpdateStore)
			store.DELETE("/", controller.DeleteStore)
			//菜品管理
			store.POST("/dishes", controller.NewDishes)
			store.GET("/dishes", controller.GetDishes)
			//菜品详情管理
			dish := store.Group("/dishes/:dishId")
			{
				dish.GET("/", controller.GetADishes)
				dish.PUT("/", controller.UpdateADishes)
				dish.DELETE("/", controller.DeleteADishes)
				//菜品标签管理
				dish.POST("/tags", controller.AddTags)
				dish.GET("/tags", controller.GetTags)
				dish.PUT("/tags", controller.ChooseTags)
				dish.DELETE("/tags/:tagId", controller.ChooseTags)
			}
		}
	}

	user := router.Group("/api/user")
	user.POST("/register", controller.UserRegister)
	user.POST("/login", controller.UserLogin)
	user.GET("/search", controller.SearchHandler)
	user.GET("/stores/:storeId", controller.AStore)
	user.GET("/stores/:storeId/dishes/:dishId", utils.AuthMiddleware(), controller.DishHandler)
	user.POST("/stores/:storeId/dishes/:dishId/rating", utils.AuthMiddleware())
	return router
}
