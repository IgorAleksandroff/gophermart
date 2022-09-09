package entity

type User struct {
	UserID    *int64
	Login     string `json:"login"`
	Password  string `json:"password"`
	Current   float64
	Withdrawn float64
}

type Balance struct {
	Login     *string
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
