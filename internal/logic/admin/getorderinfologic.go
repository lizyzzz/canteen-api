package admin

import (
	"context"

	"canteen-api/internal/biz"
	"canteen-api/internal/dao"
	"canteen-api/internal/svc"
	"canteen-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrderInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderInfoLogic {
	return &GetOrderInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderInfoLogic) GetOrderInfo(req *types.AdminOrderInfoReq) (resp *types.AdminOrderInfoResp, err error) {

	dao := dao.NewDao(l.ctx, l.svcCtx)

	isAdmin, err := dao.IsAdmin(req.UserId)
	if err != nil {
		return nil, err
	}

	if !isAdmin {
		return nil, biz.ErrNotAdmin
	}

	// 查询已完成数量
	count, err := dao.CountOrderStatus("已完成")
	if err != nil {
		return nil, err
	}

	// 查询待完成的订单
	orderInfos, err := dao.GetAdminOrderPage(req.Page, req.Size)
	if err != nil {
		return nil, err
	}

	resp = &types.AdminOrderInfoResp{
		TotalNo: count,
		Orders:  orderInfos,
	}

	return resp, nil
}
