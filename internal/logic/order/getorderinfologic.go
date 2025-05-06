package order

import (
	"context"

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

func (l *GetOrderInfoLogic) GetOrderInfo(req *types.OrderInfoReq) (resp *types.OrderInfoResp, err error) {
	dao := dao.NewDao(l.ctx, l.svcCtx)

	orderInfo, err := dao.GetOrderPage(req.UserId, req.Page, req.Size)
	if err != nil {
		return nil, err
	}

	resp = &types.OrderInfoResp{
		Orders: orderInfo,
	}

	return
}
