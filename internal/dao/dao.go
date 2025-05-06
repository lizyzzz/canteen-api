package dao

import (
	"canteen-api/internal/biz"
	"canteen-api/internal/model"
	"canteen-api/internal/svc"
	"canteen-api/internal/types"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	dishKey                 = "all_dishes"
	orderPageKey            = "order_page:"
	adminOrderPageKey       = "admin_order_page:"
	orderInfoKey            = "order_Info:"
	adminOrderStausCountKey = "admin_order_count"
	roleKey                 = "user_role"
	redisLockKey            = "lock_key"
)

type Dao struct {
	ctx    context.Context
	svrCtx *svc.ServiceContext
}

func NewDao(context context.Context, svrCtx *svc.ServiceContext) *Dao {
	return &Dao{
		ctx:    context,
		svrCtx: svrCtx,
	}
}

// 获取分布式锁
func (d *Dao) AcquireLock(userId string) (bool, error) {
	start := time.Now()
	for {
		ok, err := d.svrCtx.RedisConn.SetnxExCtx(d.ctx, redisLockKey, userId, int(d.svrCtx.Config.RedisLockExpire))
		if err != nil {
			return false, biz.RedisError
		}

		if ok {
			return true, nil
		}

		if time.Since(start) > 5*time.Second {
			return false, nil
		}

		time.Sleep(time.Millisecond * 200)
	}
}

// 释放分布式锁
func (d *Dao) ReleaseLock(userId string) error {
	luaScript := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	_, err := d.svrCtx.RedisConn.EvalCtx(d.ctx, luaScript, []string{redisLockKey}, userId)
	return err
}

func (d *Dao) GetAllDishes() (map[int64]*model.Dishes, error) {
	// 1. 查询redis是否有缓存
	dishData, err := d.svrCtx.RedisConn.HgetallCtx(d.ctx, dishKey)
	if err != nil {
		return nil, biz.RedisError
	}

	// 2. 判断缓存是否命中
	dishes := make(map[int64]*model.Dishes, 0)
	if len(dishData) == 0 {
		// 2.1 缓存未命中
		// 获取数据库模型
		m := model.NewDishesModel(d.svrCtx.Conn)
		// 查询数据库
		allDishes, err := m.GetAllDishes()
		if err != nil {
			return nil, biz.DBError
		}
		// 将数据存入redis
		keyValue := make(map[string]string)
		for _, dish := range allDishes {
			dishString, _ := json.Marshal(dish)
			keyValue[strconv.FormatInt(dish.Id, 10)] = string(dishString)
			dishes[dish.Id] = dish
		}
		// 哈希键值对存储
		err = d.svrCtx.RedisConn.HmsetCtx(d.ctx, dishKey, keyValue)
		if err != nil {
			return nil, biz.RedisError
		}
		err = d.svrCtx.RedisConn.Expire(dishKey, int(d.svrCtx.Config.RedisExpire))
		if err != nil {
			return nil, biz.RedisError
		}
	} else {
		// 2.2 缓存命中
		// json 反序列化
		for _, v := range dishData {
			var dish model.Dishes
			err = json.Unmarshal([]byte(v), &dish)
			if err != nil {
				return nil, biz.RedisError
			}
			dishes[dish.Id] = &dish
		}
	}

	return dishes, nil
}

// 删除redis缓存
func (d *Dao) DelRedisKey(key []string) error {
	_, err := d.svrCtx.RedisConn.DelCtx(d.ctx, key...)
	if err != nil {
		return biz.RedisError
	}
	return nil
}

// 生成订单分页key
func (d *Dao) GetOrderPageKey(userId string, number int) string {
	return userId + "_" + orderPageKey + strconv.Itoa(number)
}

// 生成订单分页key
func (d *Dao) GetAdminOrderPageKey(number int) string {
	return adminOrderPageKey + strconv.Itoa(number)
}

func (d *Dao) GetOrderInfoKeyByOrderId(orderId int64) string {
	return orderInfoKey + strconv.FormatInt(orderId, 10)
}

func (d *Dao) GetAdminOrderStatusKeyByStatus(status string) string {
	if status == "已完成" {
		return adminOrderStausCountKey + "_completed"
	} else {
		return adminOrderStausCountKey + "_" + status
	}
}

func (d *Dao) GetUserRoleKeyByUserId(userId string) string {
	return userId + "_" + roleKey
}

// 获取订单详情分页
func (d *Dao) GetOrderPage(userId string, pageNo, pageSize int) ([]types.OrderInfo, error) {
	// flag, err := d.AcquireLock(userId)
	// if err != nil {
	// 	return nil, biz.RedisError
	// }

	// if !flag {
	// 	return nil, biz.ErrGetRedisLock
	// }

	// defer d.ReleaseLock(userId)

	// 1. 查询redis是否有一级缓存
	orderPageKey := d.GetOrderPageKey(userId, pageNo)
	orderPageId, err := d.svrCtx.RedisConn.ZrevrangeCtx(d.ctx, orderPageKey, 0, -1)
	if err != nil {
		fmt.Println("zzz")
		return nil, biz.RedisError
	}

	orderIds := make([]int64, 0)
	// 2. 判断一级缓存是否命中
	if len(orderPageId) == 0 {
		// 2.1 一级缓存未命中
		// 获取数据库模型
		m := model.NewOrdersModel(d.svrCtx.Conn)
		// 查询数据库
		orderPage, err := m.FindOrderPageByUserId(d.ctx, userId, pageNo, pageSize)
		if err != nil {
			fmt.Println("vvv")
			return nil, biz.DBError
		}

		if len(orderPage) == 0 {
			none := make([]types.OrderInfo, 0)
			return none, nil
		}

		// 将数据存入redis
		socreKey := make([]redis.Pair, 0)
		for _, order := range orderPage {
			score := order.CreateTime.UnixMilli()
			socreKey = append(socreKey, redis.Pair{
				Key:   strconv.FormatInt(order.Id, 10),
				Score: score,
			})
			orderIds = append(orderIds, order.Id)
		}

		if pageNo <= d.svrCtx.Config.RedisOrderPageNum {
			// 只缓存部分页
			_, err = d.svrCtx.RedisConn.ZaddsCtx(d.ctx, orderPageKey, socreKey...)
			if err != nil {
				fmt.Println("ddd")
				return nil, biz.RedisError
			}
			err = d.svrCtx.RedisConn.Expire(orderPageKey, int(d.svrCtx.Config.RedisExpire))
			if err != nil {
				fmt.Println("kkk")
				return nil, biz.RedisError
			}
		}
	} else {
		// 2.2 一级缓存命中
		for _, v := range orderPageId {
			orderId, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				fmt.Println("ccc")
				return nil, biz.RedisError
			}
			orderIds = append(orderIds, orderId)
		}
	}

	orderPageInfos := make([]types.OrderInfo, 0)
	// 3. 查询redis是否有二级缓存
	for i := 0; i < len(orderIds); i++ {

		orderInfoKey := d.GetOrderInfoKeyByOrderId(orderIds[i])
		orderInfoData, err := d.svrCtx.RedisConn.GetCtx(d.ctx, orderInfoKey)
		if err != nil {
			fmt.Println("qqq")
			return nil, biz.RedisError
		}
		if len(orderInfoData) == 0 {
			// 3.1 二级缓存未命中
			// 为了节省网络请求, 直接请求全部orderId
			orderItemModel := model.NewOrderItemModel(d.svrCtx.Conn)
			orderidToInfo, err := orderItemModel.FindOrderInfoByOrderIds(d.ctx, orderIds[i:])
			if err != nil {
				return nil, biz.DBError
			}

			// 整理返回数据，并存入 redis
			for ; i < len(orderIds); i++ {
				info := types.OrderInfo{
					OrderId:    orderidToInfo[orderIds[i]].OrderId,
					Status:     orderidToInfo[orderIds[i]].Status,
					UserId:     strconv.FormatInt(orderidToInfo[orderIds[i]].UserId, 10),
					Username:   orderidToInfo[orderIds[i]].Username,
					CreateTime: orderidToInfo[orderIds[i]].CreateTime.Format("2006-01-02 15:04"),
					PickupTime: orderidToInfo[orderIds[i]].PickupTime.Format("2006-01-02"),
					TotalPrice: orderidToInfo[orderIds[i]].TotalPrice,
				}
				for _, it := range orderidToInfo[orderIds[i]].Items {
					info.Items = append(info.Items, types.OrderInfoItem{
						Id:    int(it.DishId),
						Name:  it.DishName,
						Count: int(it.Quantity),
					})
				}

				orderPageInfos = append(orderPageInfos, info)

				// 存入redis
				orderInfoKey = d.GetOrderInfoKeyByOrderId(orderIds[i])
				infoData, _ := json.Marshal(*(orderidToInfo[orderIds[i]]))
				d.svrCtx.RedisConn.SetexCtx(d.ctx, orderInfoKey, string(infoData), int(d.svrCtx.Config.RedisExpire))
			}
			break
		} else {
			// 3.2 二级缓存命中
			var orderInfo model.OrderInfo
			err = json.Unmarshal([]byte(orderInfoData), &orderInfo)
			if err != nil {
				fmt.Println("eee")
				return nil, biz.RedisError
			}

			info := types.OrderInfo{
				OrderId:    orderInfo.OrderId,
				Status:     orderInfo.Status,
				UserId:     strconv.FormatInt(orderInfo.UserId, 10),
				Username:   orderInfo.Username,
				CreateTime: orderInfo.CreateTime.Format("2006-01-02 15:04"),
				PickupTime: orderInfo.PickupTime.Format("2006-01-02"),
				TotalPrice: orderInfo.TotalPrice,
			}

			for _, it := range orderInfo.Items {
				info.Items = append(info.Items, types.OrderInfoItem{
					Id:    int(it.DishId),
					Name:  it.DishName,
					Count: int(it.Quantity),
				})
			}

			orderPageInfos = append(orderPageInfos, info)
		}
	}

	return orderPageInfos, nil
}

// 获取订单详情分页
func (d *Dao) GetAdminOrderPage(pageNo, pageSize int) ([]types.OrderInfo, error) {
	// 1. 查询redis是否有一级缓存
	orderPageKey := d.GetAdminOrderPageKey(pageNo)
	orderPageId, err := d.svrCtx.RedisConn.ZrevrangeCtx(d.ctx, orderPageKey, 0, -1)
	if err != nil {
		fmt.Println("zzz")
		return nil, biz.RedisError
	}

	orderIds := make([]int64, 0)
	// 2. 判断一级缓存是否命中
	if len(orderPageId) == 0 {
		// 2.1 一级缓存未命中
		// 获取数据库模型
		m := model.NewOrdersModel(d.svrCtx.Conn)
		// 查询数据库
		orderPage, err := m.FindAdminOrderPage(d.ctx, pageNo, pageSize)
		if err != nil {
			fmt.Println("vvv")
			return nil, biz.DBError
		}

		if len(orderPage) == 0 {
			none := make([]types.OrderInfo, 0)
			return none, nil
		}

		// 将数据存入redis
		socreKey := make([]redis.Pair, 0)
		for _, order := range orderPage {
			score := order.CreateTime.UnixMilli()
			socreKey = append(socreKey, redis.Pair{
				Key:   strconv.FormatInt(order.Id, 10),
				Score: score,
			})
			orderIds = append(orderIds, order.Id)
		}

		if pageNo <= d.svrCtx.Config.RedisOrderPageNum {
			// 只缓存部分页
			_, err = d.svrCtx.RedisConn.ZaddsCtx(d.ctx, orderPageKey, socreKey...)
			if err != nil {
				fmt.Println("ddd")
				return nil, biz.RedisError
			}
			err = d.svrCtx.RedisConn.Expire(orderPageKey, int(d.svrCtx.Config.RedisExpire))
			if err != nil {
				fmt.Println("kkk")
				return nil, biz.RedisError
			}
		}
	} else {
		// 2.2 一级缓存命中
		for _, v := range orderPageId {
			orderId, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				fmt.Println("ccc")
				return nil, biz.RedisError
			}
			orderIds = append(orderIds, orderId)
		}
	}

	orderPageInfos := make([]types.OrderInfo, 0)
	// 3. 查询redis是否有二级缓存
	for i := 0; i < len(orderIds); i++ {

		orderInfoKey := d.GetOrderInfoKeyByOrderId(orderIds[i])
		orderInfoData, err := d.svrCtx.RedisConn.GetCtx(d.ctx, orderInfoKey)
		if err != nil {
			fmt.Println("qqq")
			return nil, biz.RedisError
		}
		if len(orderInfoData) == 0 {
			// 3.1 二级缓存未命中
			// 为了节省网络请求, 直接请求全部orderId
			orderItemModel := model.NewOrderItemModel(d.svrCtx.Conn)
			orderidToInfo, err := orderItemModel.FindOrderInfoByOrderIds(d.ctx, orderIds[i:])
			if err != nil {
				return nil, biz.DBError
			}

			// 整理返回数据，并存入 redis
			for ; i < len(orderIds); i++ {
				info := types.OrderInfo{
					OrderId:    orderidToInfo[orderIds[i]].OrderId,
					Status:     orderidToInfo[orderIds[i]].Status,
					UserId:     strconv.FormatInt(orderidToInfo[orderIds[i]].UserId, 10),
					Username:   orderidToInfo[orderIds[i]].Username,
					CreateTime: orderidToInfo[orderIds[i]].CreateTime.Format("2006-01-02 15:04"),
					PickupTime: orderidToInfo[orderIds[i]].PickupTime.Format("2006-01-02"),
					TotalPrice: orderidToInfo[orderIds[i]].TotalPrice,
				}
				for _, it := range orderidToInfo[orderIds[i]].Items {
					info.Items = append(info.Items, types.OrderInfoItem{
						Id:    int(it.DishId),
						Name:  it.DishName,
						Count: int(it.Quantity),
					})
				}

				orderPageInfos = append(orderPageInfos, info)

				// 存入redis
				orderInfoKey = d.GetOrderInfoKeyByOrderId(orderIds[i])
				infoData, _ := json.Marshal(*(orderidToInfo[orderIds[i]]))
				d.svrCtx.RedisConn.SetexCtx(d.ctx, orderInfoKey, string(infoData), int(d.svrCtx.Config.RedisExpire))
			}
			break
		} else {
			// 3.2 二级缓存命中
			var orderInfo model.OrderInfo
			err = json.Unmarshal([]byte(orderInfoData), &orderInfo)
			if err != nil {
				fmt.Println("eee")
				return nil, biz.RedisError
			}

			info := types.OrderInfo{
				OrderId:    orderInfo.OrderId,
				Status:     orderInfo.Status,
				UserId:     strconv.FormatInt(orderInfo.UserId, 10),
				Username:   orderInfo.Username,
				CreateTime: orderInfo.CreateTime.Format("2006-01-02 15:04"),
				PickupTime: orderInfo.PickupTime.Format("2006-01-02"),
				TotalPrice: orderInfo.TotalPrice,
			}

			for _, it := range orderInfo.Items {
				info.Items = append(info.Items, types.OrderInfoItem{
					Id:    int(it.DishId),
					Name:  it.DishName,
					Count: int(it.Quantity),
				})
			}

			orderPageInfos = append(orderPageInfos, info)
		}
	}

	return orderPageInfos, nil
}

func (d *Dao) CountOrderStatus(status string) (int, error) {
	// 1. 先查缓存
	countKey := d.GetAdminOrderStatusKeyByStatus(status)

	result, err := d.svrCtx.RedisConn.GetCtx(d.ctx, countKey)
	if err != nil {
		return 0, err
	}

	var count int
	if result == "" {
		// 缓存未命中

		// 查数据库
		m := model.NewOrdersModel(d.svrCtx.Conn)
		count, err = m.CountOrderStatus(d.ctx, status)
		if err != nil {
			fmt.Println("dbdb", err)
			return 0, biz.DBError
		}

		// 写到redis
		err = d.svrCtx.RedisConn.SetexCtx(d.ctx, countKey, strconv.FormatInt(int64(count), 10), int(d.svrCtx.Config.RedisExpire))
		if err != nil {
			return 0, biz.RedisError
		}
	} else {
		// 缓存命中
		count, err = strconv.Atoi(result)
		if err != nil {
			return 0, biz.RedisError
		}
	}

	return count, nil
}

func (d *Dao) IsAdmin(userId string) (bool, error) {
	key := d.GetUserRoleKeyByUserId(userId)

	result, err := d.svrCtx.RedisConn.GetCtx(d.ctx, key)
	if err != nil {
		return false, err
	}

	var flag bool
	if result == "" {
		m := model.NewUserModel(d.svrCtx.Conn)

		userIdInt, _ := strconv.Atoi(userId)
		user, err := m.FindOne(d.ctx, int64(userIdInt))
		if err != nil {
			return false, biz.DBError
		}

		err = d.svrCtx.RedisConn.SetexCtx(d.ctx, key, user.Usertype, int(d.svrCtx.Config.RedisExpire))
		if err != nil {
			return false, biz.RedisError
		}

		flag = user.Usertype == "admin"
	} else {
		flag = result == "admin"
	}

	return flag, nil

}

func (d *Dao) OrderComplete(orderUserId string, orderId int64) (int, error) {
	m := model.NewOrdersModel(d.svrCtx.Conn)

	// 标记完成
	err := m.UpdateOrderStatus(d.ctx, orderId, "已完成")
	if err != nil {
		return 0, err
	}

	key := d.GetAdminOrderStatusKeyByStatus("已完成")
	result, err := d.svrCtx.RedisConn.IncrCtx(d.ctx, key)
	if err != nil {
		return 0, biz.RedisError
	}

	// 删除redis订单列表缓存

	// 管理员和用户都有一级缓存
	// 一级缓存
	delKey := make([]string, 0)
	for i := 1; i <= d.svrCtx.Config.RedisOrderPageNum; i++ {
		pageKey := d.GetAdminOrderPageKey(i)
		userPageKey := d.GetOrderPageKey(orderUserId, i)
		delKey = append(delKey, pageKey, userPageKey)
	}
	// 二级缓存
	orderKey := d.GetOrderInfoKeyByOrderId(orderId)
	delKey = append(delKey, orderKey)

	// 延迟 300ms 防止用户读取历史订单时内存中的旧数据覆盖新数据
	// 比使用 分布式锁 效率更高
	time.Sleep(300 * time.Millisecond)

	_, err = d.svrCtx.RedisConn.DelCtx(d.ctx, delKey...)
	if err != nil {
		return 0, biz.RedisError
	}

	return int(result), nil
}

// 更新用户余额
func (d *Dao) UpdateUserBalance(username string, deltaBalance float64) (*model.User, error) {
	m := model.NewUserModel(d.svrCtx.Conn)

	resp, err := m.UpdateBalanceByUsername(d.ctx, username, deltaBalance)
	if err != nil {
		return nil, biz.DBError
	}

	return resp, nil
}

func (d *Dao) InsertDish(dishInfo *model.Dishes) (int64, error) {
	m := model.NewDishesModel(d.svrCtx.Conn)

	resp, err := m.Insert(d.ctx, dishInfo)
	if err != nil {
		return 0, biz.DBError
	}

	lastId, _ := resp.LastInsertId()

	return lastId, nil
}

func (d *Dao) UpdateDishImgURL(dishInfo *model.Dishes) error {
	m := model.NewDishesModel(d.svrCtx.Conn)

	err := m.Update(d.ctx, dishInfo)
	if err != nil {
		return biz.DBError
	}

	// 删除缓存

	_, err = d.svrCtx.RedisConn.DelCtx(d.ctx, dishKey)
	if err != nil {
		return biz.RedisError
	}

	return nil
}
