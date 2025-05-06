package order

import (
	"net/http"

	"canteen-api/internal/biz"
	"canteen-api/internal/logic/order"
	"canteen-api/internal/svc"
	"canteen-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetOrderInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OrderInfoReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := order.NewGetOrderInfoLogic(r.Context(), svcCtx)
		resp, err := l.GetOrderInfo(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, biz.Success("获取订单信息成功", resp))
		}
	}
}
