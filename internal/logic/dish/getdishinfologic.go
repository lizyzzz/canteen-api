package dish

import (
	"context"

	"canteen-api/internal/dao"
	"canteen-api/internal/svc"
	"canteen-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

var categoryPriority = []string{"热销", "主食", "小吃", "饮品", "套餐", "汤类", "水果", "其他"}

type GetDishInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDishInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDishInfoLogic {
	return &GetDishInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDishInfoLogic) GetDishInfo() (resp *types.DishInfoResp, err error) {
	// 获取所有菜品
	dao := dao.NewDao(l.ctx, l.svcCtx)
	dishes, err := dao.GetAllDishes()

	if err != nil {
		return nil, err
	}

	resp = &types.DishInfoResp{
		Categories: make([]string, 0),
		DishList:   make([]types.DishInfo, 0),
	}

	categoryMap := make(map[string]bool)

	for _, dish := range dishes {
		categoryMap[dish.Category] = true

		dishInfo := types.DishInfo{
			Id:          int(dish.Id),
			Name:        dish.Name,
			Category:    dish.Category,
			Price:       dish.Price,
			Ingredients: dish.Ingredients,
			ImageUrl:    dish.ImageUrl,
		}
		resp.DishList = append(resp.DishList, dishInfo)
	}

	// 按照预定义的顺序添加分类
	for _, category := range categoryPriority {
		if _, ok := categoryMap[category]; !ok {
			continue
		}
		resp.Categories = append(resp.Categories, category)
	}

	return
}
