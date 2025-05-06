package account

import (
	"context"
	"strconv"
	"time"

	"canteen-api/internal/biz"
	"canteen-api/internal/model"
	"canteen-api/internal/svc"
	"canteen-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	// 1. 校验用户名和密码
	userModel := model.NewUserModel(l.svcCtx.Conn)
	user, err := userModel.FindUserByNameAndPwd(l.ctx, req.Username, req.Password)

	if err != nil {
		return nil, biz.DBError
	}
	if user == nil {
		return nil, biz.UserNameAndPwdError
	}

	// 2. 生成 token
	secret := l.svcCtx.Config.Auth.AccessSecret
	expire := l.svcCtx.Config.Auth.Expire

	token, err := biz.GetJwtToken(secret, time.Now().Unix(), expire, user.Id)
	if err != nil {
		return nil, biz.TokenError
	}

	resp = &types.LoginResp{
		UserId:   strconv.FormatInt(user.Id, 10),
		Username: user.Username,
		Usertype: user.Usertype,
		Balance:  user.Balance,
		Token:    token,
	}

	return
}
