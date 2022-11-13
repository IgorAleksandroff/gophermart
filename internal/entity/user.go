package entity

type User struct {
	Login     string  `json:"login" db:"login"`
	Password  string  `json:"password" db:"password"`
	Current   float64 `db:"current"`
	Withdrawn float64 `db:"withdrawn"`
}

type Balance struct {
	Login     string  `json:"-" db:"login"`
	Current   float64 `json:"current" db:"current"`
	Withdrawn float64 `json:"withdrawn" db:"withdrawn"`
}
