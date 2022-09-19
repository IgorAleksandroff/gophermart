package hendler

import (
	"net/http"

	"github.com/IgorAleksandroff/gophermart/internal/usecase"
	"github.com/IgorAleksandroff/gophermart/pkg/logger"
)

type handler struct {
	ordersUC usecase.Orders
	auth     usecase.Authorization
	l        *logger.Logger
}

type handlerFunc interface {
	MethodFunc(method, path string, handler http.HandlerFunc)
}

func New(
	ordersUC usecase.Orders,
	auth usecase.Authorization,
	l *logger.Logger,
) *handler {
	return &handler{
		ordersUC: ordersUC,
		auth:     auth,
		l:        l,
	}
}

func (h *handler) Register(router handlerFunc, method, path string, handler http.HandlerFunc) {
	router.MethodFunc(method, path, handler)
}
