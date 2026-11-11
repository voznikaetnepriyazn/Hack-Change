package domain

type Lead struct {
	FullName string `json:"full_name"`
	Login    string `json:"login"`
	Lead     bool   `json:"lead"`
}
