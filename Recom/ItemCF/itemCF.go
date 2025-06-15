package recommend

import (
	"Food_recommendation/Basic/dao"
	"Food_recommendation/Basic/model"
	"context"
	"math"
	"sort"
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
	recon := make(map[uint]float64)
	for _, dishID := range user[userID] {
		for relate, similar := range item[dishID] {
			recon[relate] += similar
		}
	}

	// 关键词匹配，增加推荐得分
	keyW := 1.0
	for _, dish := range allDishes {
		for _, keyword := range allSearch {
			if strings.Contains(dish.Name, keyword) {
				recon[dish.ID] += keyW
			}
		}
	}

	// 排序并返回推荐结果
	type RecommendItem struct {
		DishID uint
		Score  float64
	}

	// 根据推荐得分排序
	var recommendList []RecommendItem
	for dishID, score := range recon {
		recommendList = append(recommendList, RecommendItem{
			DishID: dishID,
			Score:  score,
		})
	}

	// 按得分从高到低排序
	sort.Slice(recommendList, func(i, j int) bool {
		return recommendList[i].Score > recommendList[j].Score
	})

	// 取Top N推荐结果
	var res []model.Dishes
	for i, item := range recommendList {
		if i >= 100 {
			break
		}
		for _, dish := range allDishes {
			if uint(dish.ID) == item.DishID {
				res = append(res, dish)
				break
			}
		}
	}

	return res, nil
}
