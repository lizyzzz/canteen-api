package admin

import (
	"net/http"

	"canteen-api/internal/biz"
	"canteen-api/internal/logic/admin"
	"canteen-api/internal/svc"
	"canteen-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func OrderCompleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OrderCompleteReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := admin.NewOrderCompleteLogic(r.Context(), svcCtx)
		resp, err := l.OrderComplete(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, biz.Success("订单已完成", resp))
		}
	}
}
