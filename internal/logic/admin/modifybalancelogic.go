package admin

import (
	"context"
	"strconv"

	"canteen-api/internal/biz"
	"canteen-api/internal/dao"
	"canteen-api/internal/svc"
	"canteen-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyBalanceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewModifyBalanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyBalanceLogic {
	return &ModifyBalanceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ModifyBalanceLogic) ModifyBalance(req *types.ModifyBalanceReq) (resp *types.ModifyBalanceResp, err error) {

	dao := dao.NewDao(l.ctx, l.svcCtx)

	isAdmin, err := dao.IsAdmin(req.AdminId)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, biz.ErrNotAdmin
	}

	user, err := dao.UpdateUserBalance(req.Username, req.Balance)
	if err != nil {
		return nil, err
	}

	resp = &types.ModifyBalanceResp{
		UserId:  strconv.FormatInt(user.Id, 10),
		Balance: user.Balance,
	}

	return resp, nil
}
