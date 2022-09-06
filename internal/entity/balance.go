package entity

type Balance struct {
	UserID int64 `json:"user_id"`
	Value  int64 `json:"balance"`
}
