package hendler

import (
	"net/http"

	"github.com/IgorAleksandroff/gophermart/internal/usecase"
)

type handler struct {
	ordersUC usecase.Orders
	auth     usecase.Authorization
}

type handlerFunc interface {
	MethodFunc(method, path string, handler http.HandlerFunc)
}

func New(
	ordersUC usecase.Orders,
	auth usecase.Authorization,
) *handler {
	return &handler{
		ordersUC: ordersUC,
		auth:     auth,
	}
}

func (h *handler) Register(router handlerFunc, method, path string, handler http.HandlerFunc) {
	router.MethodFunc(method, path, handler)
}
