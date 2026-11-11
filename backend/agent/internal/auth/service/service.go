package service

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/domain"
	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/storage"
)

var (
	ErrInvalidCredentials  = errors.New("invalid login or password")
	ErrForbidden           = errors.New("user is not allowed to access chat")
	ErrUserStorageRequired = errors.New("user storage is required")
)

// Service authenticates users and checks role-based access.
type Service struct {
	users storage.Storage
}

// New creates an authentication service backed by users.
func New(users storage.Storage) (*Service, error) {
	if users == nil {
		return nil, ErrUserStorageRequired
	}
	return &Service{users: users}, nil
}

// HashPassword hashes a plaintext password for storage.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password must not be empty")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Authenticate verifies credentials and returns a user without its password hash.
func (s *Service) Authenticate(login, password string) (domain.User, error) {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return domain.User{}, ErrInvalidCredentials
	}

	user, ok := s.users.FindByLogin(login)
	if !ok || user.PasswordHash == "" {
		return domain.User{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return domain.User{}, ErrInvalidCredentials
	}

	user.PasswordHash = ""
	return user, nil
}

// AuthenticateForChat authenticates the user and permits chat access to interns only.
func (s *Service) AuthenticateForChat(login, password string) (domain.User, error) {
	user, err := s.Authenticate(login, password)
	if err != nil {
		return domain.User{}, err
	}
	if user.Role != domain.RoleIntern {
		return domain.User{}, ErrForbidden
	}
	return user, nil
}

// FindByID returns a user without its password hash for an authenticated identity.
func (s *Service) FindByID(id string) (domain.User, bool) {
	user, ok := s.users.FindByID(id)
	if !ok {
		return domain.User{}, false
	}
	user.PasswordHash = ""
	return user, true
}
