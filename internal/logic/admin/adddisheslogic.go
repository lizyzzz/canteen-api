package admin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"

	"canteen-api/internal/biz"
	"canteen-api/internal/dao"
	"canteen-api/internal/model"
	"canteen-api/internal/svc"
	"canteen-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

const maxFileSize = 10 << 20 // 10MB

type AddDishesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddDishesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddDishesLogic {
	return &AddDishesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddDishesLogic) AddDishes(r *http.Request) (resp *types.UploadDishResp, err error) {
	/*
		struct
		  userId: 用户id
		  dishname: 菜品名字
		  price: 菜品价格
		  category: 菜品分类
		  ingredients: 菜品配料
		  imgfile: 菜品图片
	*/

	dao := dao.NewDao(l.ctx, l.svcCtx)

	_ = r.ParseMultipartForm(maxFileSize)
	userId := r.FormValue("userId")

	flag, err := dao.IsAdmin(userId)
	if err != nil {
		return nil, biz.DBError
	}
	if !flag {
		return nil, biz.ErrNotAdmin
	}

	// 解析参数
	dishName := r.FormValue("dishname")
	priceStr := r.FormValue("price")
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return nil, biz.ErrParam
	}
	category := r.FormValue("category")
	ingredients := r.FormValue("ingredients")

	dishInfo := &model.Dishes{
		Name:        dishName,
		Price:       price,
		Category:    category,
		Ingredients: ingredients,
	}

	// 插入数据库
	lastId, err := dao.InsertDish(dishInfo)
	if err != nil {
		return nil, err
	}

	dishInfo.Id = lastId

	file, _, err := r.FormFile("imgfile")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer file.Close()

	// 图片名称
	fileName := fmt.Sprintf("dish%d.jpg", lastId)

	dishInfo.ImageUrl = "/images/" + fileName
	// 保存图片
	tempFile, err := os.Create(path.Join(l.svcCtx.Config.Path, fileName))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer tempFile.Close()
	io.Copy(tempFile, file)

	// 更新dish
	err = dao.UpdateDishImgURL(dishInfo)
	if err != nil {
		return nil, err
	}

	return &types.UploadDishResp{
		OK: 0,
	}, nil
}
