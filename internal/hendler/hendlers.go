package hendler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/IgorAleksandroff/gophermart/internal/entity"
	"github.com/IgorAleksandroff/gophermart/internal/repository"
	"github.com/IgorAleksandroff/gophermart/internal/usecase"
)

const (
	authorizationHeader = "Authorization"
	userCtx             = "login"
)

func (h *handler) UserIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get(authorizationHeader)

		if header == "" {
			http.Error(w, "empty auth header", http.StatusUnauthorized)
			return
		}

		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			http.Error(w, "invalid auth header", http.StatusUnauthorized)
			return
		}

		login, err := h.auth.ParseToken(headerParts[1])
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		r.Header.Set(userCtx, login)
		next.ServeHTTP(w, r)
	})
}

func (h *handler) HandleUserRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	contentTypeHeaderValue := r.Header.Get("Content-Type")
	if !strings.Contains(contentTypeHeaderValue, "application/json") {
		h.l.Warn("unknown content-type")
		http.Error(w, "unknown content-type", http.StatusBadRequest)
		return
	}

	if r.Body == nil {
		h.l.Warn("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	newUser := entity.User{}
	reader := json.NewDecoder(r.Body)
	if err := reader.Decode(&newUser); err != nil {
		h.l.Warn(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.auth.CreateUser(ctx, newUser); err != nil {
		h.l.Warn(err.Error())
		if errors.Is(err, repository.ErrUserRegister) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := h.auth.GenerateToken(ctx, newUser.Login, newUser.Password)
	if err != nil {
		h.l.Warn(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(authorizationHeader, fmt.Sprintf("Bearer %s", token))
	w.WriteHeader(http.StatusOK)
}

func (h *handler) HandleUserLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	contentTypeHeaderValue := r.Header.Get("Content-Type")
	if !strings.Contains(contentTypeHeaderValue, "application/json") {
		http.Error(w, "unknown content-type", http.StatusBadRequest)
		return
	}

	if r.Body == nil {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	user := entity.User{}
	reader := json.NewDecoder(r.Body)
	if err := reader.Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.auth.GenerateToken(ctx, user.Login, user.Password)
	if err != nil {
		errLogin := errors.Is(err, usecase.ErrUserLogin) || errors.Is(err, repository.ErrUserLogin)
		if errLogin {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(authorizationHeader, fmt.Sprintf("Bearer %s", token))
	w.WriteHeader(http.StatusOK)
}

func (h *handler) HandlePostOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if !entity.Valid(orderNumber) {
		h.l.Warn("invalid order number: %v", orderNumber)
		http.Error(w, "invalid order number ", http.StatusUnprocessableEntity)
		return
	}

	err = h.ordersUC.SaveOrder(ctx, entity.Order{
		OrderID:   order,
		UserLogin: r.Header.Get(userCtx),
	})
	if err != nil {
		h.l.Warn(err.Error())
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
	ctx := r.Context()

	orders, err := h.ordersUC.GetOrders(ctx, r.Header.Get(userCtx))
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
	ctx := r.Context()

	user, err := h.ordersUC.GetUser(ctx, r.Header.Get(userCtx))
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
	ctx := r.Context()

	contentTypeHeaderValue := r.Header.Get("Content-Type")
	if !strings.Contains(contentTypeHeaderValue, "application/json") {
		http.Error(w, "unknown content-type", http.StatusBadRequest)
		return
	}

	if r.Body == nil {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	withdrawal := entity.OrderWithdraw{}
	reader := json.NewDecoder(r.Body)
	if err := reader.Decode(&withdrawal); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orderNumber, err := strconv.Atoi(withdrawal.OrderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if !entity.Valid(orderNumber) {
		http.Error(w, "invalid withdrawal number ", http.StatusUnprocessableEntity)
		return
	}

	withdrawal.UserLogin = r.Header.Get(userCtx)
	err = h.ordersUC.SaveWithdrawn(ctx, withdrawal)
	if err != nil {
		if errors.Is(err, usecase.ErrLowBalance) {
			http.Error(w, err.Error(), http.StatusPaymentRequired)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handler) HandleGetWithdrawals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orders, err := h.ordersUC.GetWithdrawals(ctx, r.Header.Get(userCtx))
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
