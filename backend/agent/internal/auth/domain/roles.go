package domain

type Role string

const (
	RoleIntern Role = "intern"
	RoleLead   Role = "lead"
)

type User struct {
	ID           string `json:"id"`
	FullName     string `json:"full_name"`
	Login        string `json:"login"`
	Role         Role   `json:"role"`
	PasswordHash string `json:"password_hash"`
}
