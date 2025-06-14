package recommend

import (
	"Food_recommendation/Basic/dao"
	"Food_recommendation/Basic/model"
	"context"
	"math"
	"strings"
)

func ItemCF(ctx context.Context, userID uint) ([]model.Dishes, error) {
	history, err := dao.GetUserHistory(ctx, userID)
	if err != nil {
		return nil, err
	}
	// 获取所有菜品
	allDishes, err := dao.GetAllDishes(ctx)
	if err != nil {
		return nil, err
	}
	allLike, err := dao.AllLike(ctx, userID)     // 返回用户喜欢的菜品集合
	allSearch, err := dao.AllSearch(ctx, userID) // 返回用户搜索过的关键词集合

	// 构建用户-物品矩阵
	userItemMatrix := make(map[uint][]uint)
	for _, h := range history {
		userItemMatrix[h.UserID] = append(userItemMatrix[h.UserID], h.StoreID)
	}

	// 构建物品-物品相似度矩阵
	itemSimilarityMatrix := make(map[uint]map[uint]float64)
	for _, userDishes := range userItemMatrix {
		for _, i := range userDishes {
			if itemSimilarityMatrix[i] == nil {
				itemSimilarityMatrix[i] = make(map[uint]float64)
			}
			for _, j := range userDishes {
				if i != j {
					itemSimilarityMatrix[i][j]++
				}
			}
		}
	}

	// 增加喜欢菜品的权重
	likeWeight := 2.0
	for _, likedDish := range allLike {
		for _, relatedDish := range allDishes {
			if likedDish.ID != relatedDish.ID {
				if itemSimilarityMatrix[likedDish.ID] == nil {
					itemSimilarityMatrix[likedDish.ID] = make(map[uint]float64)
				}
				itemSimilarityMatrix[likedDish.ID][relatedDish.ID] += likeWeight
			}
		}
	}

	// 计算物品相似度
	for i, relatedItems := range itemSimilarityMatrix {
		for j, count := range relatedItems {
			itemSimilarityMatrix[i][j] = count / math.Sqrt(float64(len(userItemMatrix[i])*len(userItemMatrix[j])))
		}
	}

	// 根据用户历史记录和物品相似度矩阵生成推荐列表
	recommendedDishesScore := make(map[uint]float64)
	for _, dishID := range userItemMatrix[userID] {
		for relatedDishID, similarity := range itemSimilarityMatrix[dishID] {
			recommendedDishesScore[relatedDishID] += similarity
		}
	}

	// 关键词匹配，增加推荐得分
	keywordWeight := 1.0
	for _, dish := range allDishes {
		for _, keyword := range allSearch {
			if strings.Contains(dish.Name, keyword) {
				recommendedDishesScore[dish.ID] += keywordWeight
			}
		}
	}

	// 排序并返回推荐结果
	var recommendedDishes []model.Dishes
	for dishID, _ := range recommendedDishesScore {
		for _, dish := range allDishes {
			if uint(dish.ID) == dishID {
				recommendedDishes = append(recommendedDishes, dish)
				break
			}
		}
	}

	return recommendedDishes, nil
}
