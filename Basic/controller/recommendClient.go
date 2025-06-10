package controller

import (
	gen "Food_recommendation/Recom/proto/gen"
	"Food_recommendation/utils"
	"context"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"strconv"
	"time"
)

func HandleItemCFRecommend(c *gin.Context) {
	id, _ := strconv.Atoi(utils.ParseSet(c))
	from, _ := strconv.Atoi(c.Query("from"))
	to, _ := strconv.Atoi(c.Query("to"))

	// 创建微服务客户端连接
	conn, err := grpc.Dial(
		"localhost:8088", // ItemCF服务地址和端口
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Printf("Failed to connect to ItemCF service: %v", err)
		c.JSON(500, gin.H{"error": "Failed to connect to recommendation service"})
		return
	}
	defer conn.Close()

	// 创建客户端
	client := gen.NewRecommendServiceClient(conn)

	// 准备请求
	itemCFReq := &gen.DishRecommendRequest{
		UserID: uint32(id),
		From:   uint32(from),
		To:     uint32(to),
	}

	// 设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 调用微服务
	resp, err := client.DishRecommend(ctx, itemCFReq)
	if err != nil {
		log.Printf("ItemCF recommendation failed: %v", err)
		c.JSON(500, gin.H{"error": "Recommendation service error"})
		return
	}

	// 返回成功响应
	c.JSON(200, gin.H{
		"status": "success",
		"data":   resp.Recommendations,
	})
}
