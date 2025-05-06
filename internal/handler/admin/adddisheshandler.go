package admin

import (
	"net/http"

	"canteen-api/internal/biz"
	"canteen-api/internal/logic/admin"
	"canteen-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func AddDishesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := admin.NewAddDishesLogic(r.Context(), svcCtx)
		resp, err := l.AddDishes(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, biz.Success("添加成功", resp))
		}
	}
}
