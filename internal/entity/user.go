package entity

type User struct {
	Login    string `json:"login"`
	password string `json:"password"`
	UserID   *int64 `json:"user_id,omitempty"`
}
