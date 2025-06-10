package main

import (
	gen "Food_recommendation/Recom/proto/gen"
	recommend "Food_recommendation/Recom/server"
	"context"
	"log"
	"net"

	"Food_recommendation/Basic/dao"
	"Food_recommendation/Basic/model"
	"google.golang.org/grpc"
)

// RecommendServer 实现 RecommendService 接口
type RecommendServer struct {
	gen.UnimplementedRecommendServiceServer
}

// DishRecommend 实现菜品推荐方法
func (s *RecommendServer) DishRecommend(ctx context.Context, req *gen.DishRecommendRequest) (*gen.DishRecommendResponse, error) {
	// 调用 ItemCF 函数获取推荐菜品
	recommendedDishes, err := recommend.ItemCF(ctx, uint(req.UserID))
	if err != nil {
		return nil, err
	}

	// 截取需要的推荐数量
	if len(recommendedDishes) > int(req.To) {
		recommendedDishes = recommendedDishes[req.From:req.To]
	} else if len(recommendedDishes) > int(req.From) {
		recommendedDishes = recommendedDishes[req.From:]
	} else {
		recommendedDishes = []model.Dishes{}
	}

	response := &gen.DishRecommendResponse{}
	for _, dish := range recommendedDishes {
		// 获取店铺名称
		store, err := dao.SearchStore(ctx, dish.StoreID)
		if err != nil {
			return nil, err
		}

		recommendation := &gen.DishRecommendation{
			DishID:    uint32(dish.ID),
			DishName:  dish.Name,
			StoreID:   uint32(dish.StoreID),
			StoreName: store.Name,
			ImageURL:  dish.ImageURL,
			AvgRating: float32(dish.AvgRating),
			Price:     dish.Price.String(),
			Desc:      dish.Desc,
			Available: dish.Available,
		}
		response.Recommendations = append(response.Recommendations, recommendation)
	}

	return response, nil
}

func main() {
	// 初始化数据库
	dao.InitDB()

	// 创建 gRPC 服务器
	lis, err := net.Listen("tcp", ":8088")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	// 注册服务
	gen.RegisterRecommendServiceServer(s, &RecommendServer{})

	log.Println("Starting gRPC server on port 50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
