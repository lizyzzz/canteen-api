package biz

// 自定义返回结果
type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
	Data any    `json:"data"`
}

func Success(msg string, data any) *Result {
	return &Result{
		Code: OK,
		Msg:  msg,
		Data: data,
	}
}

func Fail(err *Error) *Result {
	return &Result{
		Code: err.Code,
		Msg:  err.Msg,
		Data: nil,
	}
}
