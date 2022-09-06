package hendler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/IgorAleksandroff/gophermart.git/internal/entity"
	"github.com/IgorAleksandroff/gophermart.git/internal/usecase"
)

func (h *handler) HandlePostOrders(w http.ResponseWriter, r *http.Request) {
	contentTypeHeaderValue := r.Header.Get("Content-Type")
	if !strings.Contains(contentTypeHeaderValue, "application/json") {
		http.Error(w, "unknown content-type", http.StatusBadRequest)
		return
	}

	if r.Body == nil {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	orderNumber, err := strconv.Atoi(string(b))
	if err != nil || !entity.Valid(orderNumber) {
		http.Error(w, "invalid order number "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// todo: поход за бонусами
	// todo: проверка пользователя
	err = h.ordersUC.SaveOrder(entity.Order{
		OrderID:    int64(orderNumber),
		UserID:     0,
		UploadedAt: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		if errors.Is(err, usecase.ErrExistOrderByThisUser) {
			http.Error(w, err.Error(), http.StatusOK)
			return
		}
		if errors.Is(err, usecase.ErrExistOrderByAnotherUser) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *handler) HandleGetOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.ordersUC.GetOrders()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		http.Error(w, "empty slice", http.StatusNoContent)
		return
	}

	buf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(buf)
	err = jsonEncoder.Encode(orders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}
