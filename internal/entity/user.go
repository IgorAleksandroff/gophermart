package entity

type User struct {
	UserID    *int64
	Login     string
	Password  string
	Current   int64
	Withdrawn int64
}

type Balance struct {
	userID    int64
	Current   int64 `json:"current"`
	Withdrawn int64 `json:"withdrawn"`
}
