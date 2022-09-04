package hendler

import (
	"net/http"

	"github.com/IgorAleksandroff/gophermart.git/internal/usecase"
)

type handler struct {
	ordersUC usecase.Orders
}

type handlerFunc interface {
	MethodFunc(method, path string, handler http.HandlerFunc)
}

func New(
	ordersUC usecase.Orders,
) *handler {
	return &handler{
		ordersUC: ordersUC,
	}
}

func (h *handler) Register(router handlerFunc, method, path string, handler http.HandlerFunc) {
	router.MethodFunc(method, path, handler)
}
