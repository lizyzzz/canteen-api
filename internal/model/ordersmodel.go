package model

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ OrdersModel = (*customOrdersModel)(nil)

type (
	// OrdersModel is an interface to be customized, add more methods here,
	// and implement the added methods in customOrdersModel.
	OrdersModel interface {
		ordersModel
		withSession(session sqlx.Session) OrdersModel
		FindOrderPageByUserId(ctx context.Context, userId string, pageNo, pageSize int) ([]*Orders, error)
		FindAdminOrderPage(ctx context.Context, pageNo, pageSize int) ([]*Orders, error)
		CountOrderStatus(ctx context.Context, status string) (int, error)
		UpdateOrderStatus(ctx context.Context, orderId int64, status string) error
	}

	customOrdersModel struct {
		*defaultOrdersModel
	}
)

// NewOrdersModel returns a model for the database table.
func NewOrdersModel(conn sqlx.SqlConn) OrdersModel {
	return &customOrdersModel{
		defaultOrdersModel: newOrdersModel(conn),
	}
}

func (m *customOrdersModel) withSession(session sqlx.Session) OrdersModel {
	return NewOrdersModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customOrdersModel) FindOrderPageByUserId(ctx context.Context, userId string, pageNo, pageSize int) ([]*Orders, error) {
	// 1. 获取分页数据
	query := fmt.Sprintf("select %s from %s where user_id = ? order by create_time desc limit ?, ?", ordersRows, m.table)

	id, _ := strconv.Atoi(userId)

	var resp []*Orders
	err := m.conn.QueryRowsCtx(ctx, &resp, query, id, (pageNo-1)*pageSize, pageSize)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound, sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}

}

func (m *customOrdersModel) FindAdminOrderPage(ctx context.Context, pageNo, pageSize int) ([]*Orders, error) {
	// 1. 获取分页数据
	query := fmt.Sprintf("select %s from %s where status = ? order by create_time desc limit ?, ?", ordersRows, m.table)

	var resp []*Orders
	err := m.conn.QueryRowsCtx(ctx, &resp, query, "待取餐", (pageNo-1)*pageSize, pageSize)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound, sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}

}

type CountDB struct {
	Cnt int `db:"cnt"`
}

func (m *customOrdersModel) CountOrderStatus(ctx context.Context, status string) (int, error) {
	query := fmt.Sprintf("select count(*) as cnt from %s where status = ?", m.table)

	var resp CountDB
	err := m.conn.QueryRowCtx(ctx, &resp, query, status)

	switch err {
	case nil:
		return resp.Cnt, nil
	case sqlx.ErrNotFound, sql.ErrNoRows:
		return 0, nil
	default:
		return 0, err
	}
}

func (m *customOrdersModel) UpdateOrderStatus(ctx context.Context, orderId int64, status string) error {
	query := fmt.Sprintf("update %s set `status` = ? where id = ?", m.table)

	err := m.conn.QueryRowCtx(ctx, nil, query, status, orderId)

	switch err {
	case nil:
		return nil
	case sqlx.ErrNotFound, sql.ErrNoRows:
		return nil
	default:
		return err
	}
}
