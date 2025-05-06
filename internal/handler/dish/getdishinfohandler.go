package dish

import (
	"net/http"

	"canteen-api/internal/biz"
	"canteen-api/internal/logic/dish"
	"canteen-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetDishInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := dish.NewGetDishInfoLogic(r.Context(), svcCtx)
		resp, err := l.GetDishInfo()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, biz.Success("获取菜品成功", resp))
		}
	}
}
