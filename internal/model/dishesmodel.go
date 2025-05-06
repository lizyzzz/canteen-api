package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ DishesModel = (*customDishesModel)(nil)

type (
	// DishesModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDishesModel.
	DishesModel interface {
		dishesModel
		withSession(session sqlx.Session) DishesModel
		GetAllDishes() ([]*Dishes, error)
	}

	customDishesModel struct {
		*defaultDishesModel
	}
)

// NewDishesModel returns a model for the database table.
func NewDishesModel(conn sqlx.SqlConn) DishesModel {
	return &customDishesModel{
		defaultDishesModel: newDishesModel(conn),
	}
}

func (m *customDishesModel) withSession(session sqlx.Session) DishesModel {
	return NewDishesModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customDishesModel) GetAllDishes() ([]*Dishes, error) {
	query := "SELECT id, name, price, category, ingredients, image_url FROM dishes"
	var resp []*Dishes
	err := m.conn.QueryRows(&resp, query)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
