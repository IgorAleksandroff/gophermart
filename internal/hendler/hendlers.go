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
	if !strings.Contains(contentTypeHeaderValue, "text/plain") {
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

	order := string(b)
	orderNumber, err := strconv.Atoi(order)
	if err != nil || !entity.Valid(orderNumber) {
		http.Error(w, "invalid order number "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// todo: проверка пользователя
	err = h.ordersUC.SaveOrder(entity.Order{
		OrderID:    order,
		UserLogin:  "testuser",
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func (h *handler) HandleGetBalance(w http.ResponseWriter, r *http.Request) {
	// todo: добавить пользователя
	user, err := h.ordersUC.GetUser("testuser")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	balance := entity.Balance{
		Current:   user.Current,
		Withdrawn: user.Withdrawn,
	}

	buf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(buf)
	err = jsonEncoder.Encode(balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func (h *handler) HandlePostBalanceWithdraw(w http.ResponseWriter, r *http.Request) {
	// todo: проверка пользователя
	contentTypeHeaderValue := r.Header.Get("Content-Type")
	if !strings.Contains(contentTypeHeaderValue, "application/json") {
		http.Error(w, "unknown content-type", http.StatusBadRequest)
		return
	}

	if r.Body == nil {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	withdrawn := entity.OrderWithdraw{}
	reader := json.NewDecoder(r.Body)
	if err := reader.Decode(&withdrawn); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orderNumber, err := strconv.Atoi(withdrawn.OrderID)
	if err != nil || !entity.Valid(orderNumber) {
		http.Error(w, "invalid order number "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	withdrawn.UserLogin = "testuser"
	err = h.ordersUC.SaveWithdrawn(withdrawn)
	if err != nil {
		if errors.Is(err, usecase.ErrLowBalance) {
			http.Error(w, err.Error(), http.StatusPaymentRequired)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *handler) HandleGetWithdrawals(w http.ResponseWriter, r *http.Request) {
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
