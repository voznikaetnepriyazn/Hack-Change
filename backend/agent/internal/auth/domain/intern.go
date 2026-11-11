package domain

type Intern struct {
	FullName string `json:"full_name"`
	Login    string `json:"login"`
	Intern   bool   `json:"intern"`
}
