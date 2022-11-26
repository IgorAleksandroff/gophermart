package entity

type Order struct {
	OrderID    string  `db:"order_id"`
	UserLogin  string  `db:"login"`
	Status     string  `db:"status"`
	Accrual    float64 `db:"accrual"`
	UploadedAt string  `db:"uploaded_at"`
}

type OrderWithdraw struct {
	OrderID     string  `json:"order" db:"order_id"`
	UserLogin   string  `db:"login"`
	Value       float64 `json:"sum" db:"value"`
	ProcessedAt string  `json:"processed_at,omitempty" db:"processed_at"`
}

type Orders struct {
	OrderID    string  `json:"number" db:"order_id"`
	Status     string  `json:"status" db:"status"`
	Accrual    float64 `json:"accrual,omitempty" db:"accrual"`
	UploadedAt string  `json:"uploaded_at" db:"uploaded_at"`
}

type Accrual struct {
	OrderID string   `json:"order"`
	Status  string   `json:"status,omitempty"`
	Accrual *float64 `json:"accrual,omitempty"`
}

// Valid check number is valid or not based on Luhn algorithm
func Valid(number int) bool {
	return (number%10+checksum(number/10))%10 == 0
}

func checksum(number int) int {
	var luhn int

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 { // even
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}
	return luhn % 10
}
