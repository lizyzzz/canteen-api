package admin

import (
	"net/http"

	"canteen-api/internal/biz"
	"canteen-api/internal/logic/admin"
	"canteen-api/internal/svc"
	"canteen-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ModifyBalanceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ModifyBalanceReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := admin.NewModifyBalanceLogic(r.Context(), svcCtx)
		resp, err := l.ModifyBalance(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, biz.Success("修改成功", resp))
		}
	}
}
