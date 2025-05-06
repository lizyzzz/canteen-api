package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ OrderItemModel = (*customOrderItemModel)(nil)

type (
	// OrderItemModel is an interface to be customized, add more methods here,
	// and implement the added methods in customOrderItemModel.
	OrderItemModel interface {
		orderItemModel
		withSession(session sqlx.Session) OrderItemModel
		FindOrderInfoByOrderIds(ctx context.Context, orderId []int64) (map[int64]*OrderInfo, error)
	}

	customOrderItemModel struct {
		*defaultOrderItemModel
	}
)

// NewOrderItemModel returns a model for the database table.
func NewOrderItemModel(conn sqlx.SqlConn) OrderItemModel {
	return &customOrderItemModel{
		defaultOrderItemModel: newOrderItemModel(conn),
	}
}

func (m *customOrderItemModel) withSession(session sqlx.Session) OrderItemModel {
	return NewOrderItemModel(sqlx.NewSqlConnFromSession(session))
}

type Item struct {
	DishId    int64   `db:"dish_id"`
	DishName  string  `db:"dish_name"`
	UnitPrice float64 `db:"unit_price"`
	Quantity  uint64  `db:"quantity"`
	Subtotal  float64 `db:"subtotal"`
}

type OrderInfoDB struct {
	Item
	Id         int64     `db:"id"`
	UserId     int64     `db:"user_id"`
	Username   string    `db:"username"`
	Status     string    `db:"status"`
	CreateTime time.Time `db:"create_time"`
	PickupTime time.Time `db:"pickup_time"`
	TotalPrice float64   `db:"total_price"`
}

type OrderInfo struct {
	OrderId    int64
	UserId     int64
	Username   string
	Status     string
	CreateTime time.Time
	PickupTime time.Time
	TotalPrice float64
	Items      []*Item
}

// 查询订单详情
func (m *customOrderItemModel) FindOrderInfoByOrderIds(ctx context.Context, orderId []int64) (map[int64]*OrderInfo, error) {
	// 查询所有orderid的详情 (联合查询)
	query := `select o.id, o.user_id, o.username, o.status, o.create_time, o.pickup_time, o.total_price, 
			  item.dish_id, item.dish_name, item.unit_price, item.quantity, item.subtotal
			  from orders o
			  left join order_item item 
			  on o.id = item.order_id
			  where o.id in (`
	for i := 0; i < len(orderId); i++ {
		query += fmt.Sprintf("%d", orderId[i])
		if i != len(orderId)-1 {
			query += ","
		}
	}
	query += ")"

	var resp []*OrderInfoDB
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	if err != nil {
		fmt.Println(err)
		switch err {
		case sqlx.ErrNotFound, sql.ErrNoRows:
			return nil, nil
		default:
			return nil, err
		}
	}

	// 将结果转换为map
	orderInfoResp := make(map[int64]*OrderInfo)
	for _, item := range resp {
		var oi *OrderInfo
		oi, ok := orderInfoResp[item.Id]
		if !ok {
			oi = &OrderInfo{
				OrderId:    item.Id,
				UserId:     item.UserId,
				Username:   item.Username,
				Status:     item.Status,
				CreateTime: item.CreateTime,
				PickupTime: item.PickupTime,
				TotalPrice: item.TotalPrice,
			}
			orderInfoResp[item.Id] = oi
		}

		oi.Items = append(oi.Items, &Item{
			DishId:    item.DishId,
			DishName:  item.DishName,
			UnitPrice: item.UnitPrice,
			Quantity:  item.Quantity,
			Subtotal:  item.Subtotal,
		})
	}
	return orderInfoResp, nil
}
