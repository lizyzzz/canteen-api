package biz

const OK = 200

var (
	DBError             = NewError(10000, "数据库错误")
	AlreadyRegister     = NewError(10100, "用户已注册")
	UserNameAndPwdError = NewError(10101, "用户名或密码错误")
	TokenError          = NewError(10102, "Token错误")
	RedisError          = NewError(10103, "Redis错误")
	InvalidInviteCode   = NewError(10104, "邀请码无效")
	ErrDishNotFound     = NewError(10105, "菜品不存在")
	ErrUserNotFound     = NewError(10105, "用户不存在")
	ErrOrderParam       = NewError(10106, "订单参数错误")
	ErrBalanceNotEnough = NewError(10107, "余额不足")
	ErrNotAdmin         = NewError(10108, "非管理员")
	ErrGetRedisLock     = NewError(10109, "获取锁失败")
	ErrParam            = NewError(10110, "参数错误")
	ErrCommonError      = NewError(10111, "未知错误")
)
