package dao

import (
	"Food_recommendation/Basic/model"
	"Food_recommendation/utils"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
)

func CreateUser(ctx context.Context, u model.User) error {
	var count int64
	if err := DB.WithContext(ctx).Model(&model.User{}).
		Where("username = ?", u.Username).Count(&count).Error; err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	if count > 0 {
		return errors.New("username already exists")
	}
	password, _ := utils.Crypto(u.Password)
	u.Password = password
	if err := DB.Create(&u).Error; err != nil {
		log.Println("Create user Error!")
		return err
	}
	log.Println("Create user Success!")
	return nil
}
func UserLogin(ctx context.Context, u model.User) (model.User, error) {
	inputPass, _ := utils.Crypto(u.Password)
	if err := DB.WithContext(ctx).Where("username = ? AND password = ?", u.Username, inputPass).First(&u).Error; err != nil {
		log.Println("Password is incorrect!")
		return u, errors.New("password is incorrect")
	}
	fmt.Println(u)
	return u, nil
}

// func History(ctx context.Context, u model.User) error {
//
// }
func UserSearch(ctx context.Context, keyword string, uid uint) ([]model.ShowMerchant, error) {
	trimmedKeyword := strings.TrimSpace(keyword)

	// 处理空关键词 - 返回随机20个菜品
	if trimmedKeyword == "" || uid == 0 {
		var results []model.ShowMerchant
		err := DB.WithContext(ctx).
			Table("dishes d").
			Select(`
				d.id AS dishes_id,
                d.image_url AS img,
                d.name AS dishes_name,
				d.like_num AS likenum,
                s.name AS store_name,
				s.address AS store_address,
                FORMAT(d.avg_rating, 1) AS rating,
                CONCAT("/store/", s.id) AS link
            `).
			Joins("JOIN stores s ON d.store_id = s.id").
			Order("RAND()"). // 使用数据库的随机排序功能
			Limit(20).       // 限制返回20条记录
			Scan(&results).Error

		if err != nil {
			return nil, fmt.Errorf("get random dishes failed: %w", err)
		}

		return results, nil
	}
	DB.Model(model.Search{}).Create(&model.Search{
		UserID: uid,
		Key:    keyword,
	})
	// 关键词不为空 - 执行原有的搜索逻辑
	keywords := strings.Fields(trimmedKeyword)
	conditions := make([]string, len(keywords))
	args := make([]interface{}, len(keywords))

	for i, word := range keywords {
		conditions[i] = "d.name LIKE ?"
		args[i] = "%" + word + "%"
	}

	whereCondition := strings.Join(conditions, " OR ")
	orderCondition := buildOrderCondition(keyword, keywords)

	var results []model.ShowMerchant
	err := DB.WithContext(ctx).
		Table("dishes d").
		Select(`
            d.id AS dishes_id,
            d.image_url AS img,
            d.name AS dishes_name,
			d.like_num AS likenum,
            s.name AS store_name,
			s.address AS store_address,
            FORMAT(d.avg_rating, 1) AS rating,
            CONCAT("/store/", s.id) AS link
        `).
		Joins("JOIN stores s ON d.store_id = s.id").
		Where(whereCondition, args...).
		Order(orderCondition).
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("search dishes failed: %w", err)
	}
	return results, nil
}

func buildOrderCondition(keyword string, keywords []string) string {
	allKeywordsPattern := "%" + strings.Join(keywords, "%") + "%"
	orderSQL := fmt.Sprintf(`
        CASE 
            WHEN d.name LIKE '%s' THEN 3  
            WHEN d.name LIKE '%s' THEN 2  
            ELSE 1                       
        END DESC,
        d.avg_rating DESC
    `,
		strings.ReplaceAll(keyword, "'", "''"),
		strings.ReplaceAll(allKeywordsPattern, "'", "''"))
	return orderSQL
}

func AddHistory(ctx context.Context, uid uint, SID uint) error {
	// 创建 History 实例
	history := model.History{
		UserID:  uid,
		StoreID: SID,
	}
	// 使用上下文关联的 DB 执行插入（可自动处理连接池）
	if err := DB.WithContext(ctx).Create(&history).Error; err != nil {
		return fmt.Errorf("failed to add history record: %w", err)
	}
	return nil
}
func AllSearch(ctx context.Context, uid uint) ([]string, error) {
	var results []string
	err := DB.Debug().Model(model.Search{}).WithContext(ctx).Select("key").Where("user_id = ?", uid).Order("created_at DESC").Limit(20).Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("get search records failed: %w", err)
	}
	return results[:20], nil
}
func AllLike(ctx context.Context, uid uint) ([]model.Dishes, error) {
	var results []model.Dishes
	var likes []struct {
		DishID string `gorm:"column:dish_id"`
	}

	// 查询用户喜欢的菜品ID
	err := DB.WithContext(ctx).Model(&model.Like{}).
		Select("dish_id").
		Where("user_id = ?", uid).
		Order("created_at DESC").
		Find(&likes).Error
	if err != nil {
		return nil, fmt.Errorf("get likes records failed: %w", err)
	}

	// 收集所有菜品ID
	var dishIDs []string
	for _, like := range likes {
		dishIDs = append(dishIDs, like.DishID)
	}

	// 批量查询菜品信息
	if len(dishIDs) > 0 {
		err = DB.WithContext(ctx).
			Select("dishes.id", "dishes.store_id", "dishes.name", "dishes.price", "dishes.desc", "dishes.image_url", "dishes.available", "dishes.like_num", "dishes.avg_rating").
			Preload("Tags", "tags.name != ''"). // 预加载非空标签
			Where("dishes.id IN ?", dishIDs).
			Find(&results).Error
		if err != nil {
			return nil, fmt.Errorf("get dishes failed: %w", err)
		}
	}

	return results, nil
}
