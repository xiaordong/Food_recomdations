package main

import (
	"Food_recommendation/Basic/dao"
	"Food_recommendation/Basic/router"
)

func main() {
	dao.InitDB()
	r := router.InitRouter()
	r.Run(":6001")
}
