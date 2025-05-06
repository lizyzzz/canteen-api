package order

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"canteen-api/internal/biz"
	"canteen-api/internal/dao"
	"canteen-api/internal/model"
	"canteen-api/internal/svc"
	"canteen-api/internal/types"

	"github.com/sony/sonyflake"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var sf *sonyflake.Sonyflake

func init() {
	var st sonyflake.Settings
	sf = sonyflake.NewSonyflake(st)
	if sf == nil {
		panic("sonyflake not created")
	}
}

type PayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PayOrderLogic {
	return &PayOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayOrderLogic) PayOrder(req *types.PayOrderReq) (resp *types.PayOrderResp, err error) {
	// 1. 校验参数
	// 计算价格
	dao := dao.NewDao(l.ctx, l.svcCtx)
	dishes, err := dao.GetAllDishes()
	if err != nil {
		return nil, err
	}

	totalPrice := 0.0
	for _, item := range req.Items {
		dish, ok := dishes[int64(item.Id)]
		if !ok {
			return nil, biz.ErrDishNotFound
		}
		totalPrice += dish.Price * float64(item.Count)
	}
	totalPrice = 0.0
	if math.Abs(totalPrice-req.TotalPrice) >= 0.1 {
		return nil, biz.ErrOrderParam
	}

	// 2. 开启事务
	var user model.User
	err = l.svcCtx.Conn.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		userId, _ := strconv.Atoi(req.UserId)
		fmt.Println("userId", userId)
		// 2.1 查询用户余额
		dberr := session.QueryRowCtx(ctx, &user, "select * from `user` where id = ?", userId)
		if dberr != nil {
			if dberr == sqlx.ErrNotFound {
				return biz.ErrUserNotFound
			}
			fmt.Println("a", dberr)
			return biz.DBError
		}
		// 余额不足
		if user.Balance < req.TotalPrice {
			return biz.ErrBalanceNotEnough
		}

		// 2.2 插入订单
		order_no, e := sf.NextID()
		if e != nil {
			fmt.Println("b")
			return biz.ErrCommonError
		}

		pTime, e := time.Parse("2006-01-02", req.PickupTime)
		if e != nil {
			return biz.ErrOrderParam
		}
		order := &model.Orders{
			UserId:     int64(userId),
			Username:   user.Username,
			OrderNo:    order_no,
			TotalPrice: req.TotalPrice,
			PickupTime: pTime,
			CreateTime: time.Now(),
		}
		// 插入订单
		result, dberr := session.ExecCtx(ctx, "insert into `orders` (user_id, username, order_no, total_price, pickup_time, create_time) values (?, ?, ?, ?, ?, ?)",
			order.UserId, order.Username, order.OrderNo, order.TotalPrice, order.PickupTime, order.CreateTime)
		if dberr != nil {
			fmt.Println("c")
			return biz.DBError
		}

		// 2.3 插入订单详情
		// 构造批量插入的 SQL 语句
		lastInsertID, err := result.LastInsertId()
		if err != nil {
			return biz.DBError
		}
		query := `INSERT INTO order_item (order_id, dish_id, dish_name, unit_price, quantity) VALUES `
		var args []interface{}
		var values []string
		for _, item := range req.Items {
			values = append(values, "(?, ?, ?, ?, ?)")
			args = append(args, lastInsertID, int64(item.Id), item.Name, item.Price, item.Count)
		}
		query += strings.Join(values, ", ")
		_, dberr = session.ExecCtx(ctx, query, args...)
		if dberr != nil {
			fmt.Println("d")
			return biz.DBError
		}

		// 2.4 扣除用户余额
		balance := user.Balance - req.TotalPrice
		_, dberr = session.ExecCtx(ctx, "update `user` set balance = ? where id = ?", balance, int64(userId))
		if dberr != nil {
			fmt.Println("d")
			return biz.DBError
		}

		// 2.5 删除 redis 中历史订单分页的键
		orderKey := make([]string, 0)
		for i := 1; i <= l.svcCtx.Config.RedisOrderPageNum; i++ {
			orderKey = append(orderKey, dao.GetOrderPageKey(req.UserId, i), dao.GetAdminOrderPageKey(i))
		}
		rdsErr := dao.DelRedisKey(orderKey)
		if rdsErr != nil {
			return biz.RedisError
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	resp = &types.PayOrderResp{
		Balance: user.Balance - req.TotalPrice,
	}

	return resp, nil
}
