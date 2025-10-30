package model

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Login    string `json:"login"`
	Password string `json:"-"`
}
