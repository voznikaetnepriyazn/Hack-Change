package storage

import "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/domain"

// Storage provides read access to users needed by authentication.
type Storage interface {
	FindByLogin(login string) (domain.User, bool)
	FindByID(id string) (domain.User, bool)
}
