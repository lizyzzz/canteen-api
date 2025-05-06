package account

import (
	"context"
	"time"

	"canteen-api/internal/biz"
	"canteen-api/internal/model"
	"canteen-api/internal/svc"
	"canteen-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

const (
	UserTypeAdmin = "admin"
	UserTypeUser  = "user"
)

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	// 注册逻辑

	// 0. 校验邀请码(相当于白名单)
	var usertype string
	if req.Invitecode == l.svcCtx.Config.AdminInviteCode {
		usertype = UserTypeAdmin
	} else if req.Invitecode == l.svcCtx.Config.UserInviteCode {
		usertype = UserTypeUser
	} else {
		return nil, biz.InvalidInviteCode
	}

	// 1. 根据用户名查询是否已经注册
	userModel := model.NewUserModel(l.svcCtx.Conn)
	user, err := userModel.FindUserByName(l.ctx, req.Username)

	if err != nil {
		return nil, biz.DBError
	}

	if user != nil {
		return nil, biz.AlreadyRegister
	}

	// 2. 用户没注册则插入
	_, err = userModel.Insert(l.ctx, &model.User{
		Username:     req.Username,
		Password:     req.Password,
		Usertype:     usertype,
		RegisterTime: time.Now(),
	})

	if err != nil {
		return nil, biz.DBError
	}

	return
}
