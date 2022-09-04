package entity

type Order struct {
	ID    int64  `json:"id"`
	Value *int64 `json:"value,omitempty"`
}
