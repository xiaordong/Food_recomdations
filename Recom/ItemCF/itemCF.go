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
	user := make(map[uint][]uint)
	for _, h := range history {
		user[h.UserID] = append(user[h.UserID], h.StoreID)
	}

	// 构建物品-物品相似度矩阵
	item := make(map[uint]map[uint]float64)
	for _, userDishes := range user {
		for _, i := range userDishes {
			if item[i] == nil {
				item[i] = make(map[uint]float64)
			}
			for _, j := range userDishes {
				if i != j {
					item[i][j]++
				}
			}
		}
	}

	// 增加喜欢菜品的权重
	like := 2.0
	for _, likedDish := range allLike {
		for _, relatedDish := range allDishes {
			if likedDish.ID != relatedDish.ID {
				if item[likedDish.ID] == nil {
					item[likedDish.ID] = make(map[uint]float64)
				}
				item[likedDish.ID][relatedDish.ID] += like
			}
		}
	}

	// 计算物品相似度
	for i, similar := range item {
		for j, count := range similar {
			item[i][j] = count / math.Sqrt(float64(len(user[i])*len(user[j])))
		}
	}

	// 根据用户历史记录和物品相似度矩阵生成推荐列表
	recommend := make(map[uint]float64)
	for _, dishID := range user[userID] {
		for relatedDishID, similarity := range item[dishID] {
			recommend[relatedDishID] += similarity
		}
	}

	// 关键词匹配，增加推荐得分
	keywordWeight := 1.0
	for _, dish := range allDishes {
		for _, keyword := range allSearch {
			if strings.Contains(dish.Name, keyword) {
				recommend[dish.ID] += keywordWeight
			}
		}
	}

	// 排序并返回推荐结果
	var recommendedDishes []model.Dishes
	for dishID, _ := range recommend {
		for _, dish := range allDishes {
			if uint(dish.ID) == dishID {
				recommendedDishes = append(recommendedDishes, dish)
				break
			}
		}
	}

	return recommendedDishes, nil
}
