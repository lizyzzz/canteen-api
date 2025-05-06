package admin

import (
	"context"

	"canteen-api/internal/biz"
	"canteen-api/internal/dao"
	"canteen-api/internal/svc"
	"canteen-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderCompleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderCompleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderCompleteLogic {
	return &OrderCompleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderCompleteLogic) OrderComplete(req *types.OrderCompleteReq) (resp *types.OrderCompleteResp, err error) {
	// todo: add your logic here and delete this line
	dao := dao.NewDao(l.ctx, l.svcCtx)

	isAdmin, err := dao.IsAdmin(req.UserId)
	if err != nil {
		return nil, err
	}

	if !isAdmin {
		return nil, biz.ErrNotAdmin
	}

	completeCnt, err := dao.OrderComplete(req.OrderUserId, req.OrderId)
	if err != nil {
		return nil, err
	}

	resp = &types.OrderCompleteResp{
		TotalNo: completeCnt,
	}

	return resp, nil
}
