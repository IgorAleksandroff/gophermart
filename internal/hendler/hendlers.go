package hendler

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func (h *handler) HandlePostOrders(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
}

func (h *handler) HandleGetOrders(w http.ResponseWriter, r *http.Request) {
	orders := h.ordersUC.GetOrders()

	buf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(buf)
	err := jsonEncoder.Encode(orders)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}
