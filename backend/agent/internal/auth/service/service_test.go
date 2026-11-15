package service

import (
	"errors"
	"testing"

	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/domain"
)

type userStorage map[string]domain.User

func (s userStorage) FindByLogin(login string) (domain.User, bool) {
	user, ok := s[login]
	return user, ok
}

func (s userStorage) FindByID(id string) (domain.User, bool) {
	for _, user := range s {
		if user.ID == id {
			return user, true
		}
	}
	return domain.User{}, false
}

func newTestService(t *testing.T) *Service {
	t.Helper()

	internHash, err := HashPassword("intern-password")
	if err != nil {
		t.Fatalf("HashPassword(intern): %v", err)
	}
	leadHash, err := HashPassword("lead-password")
	if err != nil {
		t.Fatalf("HashPassword(lead): %v", err)
	}

	service, err := New(userStorage{
		"intern": {
			ID:           "user-1",
			FullName:     "Alex Intern",
			Login:        "intern",
			Role:         domain.RoleIntern,
			PasswordHash: internHash,
		},
		"lead": {
			ID:           "user-2",
			FullName:     "Sam Lead",
			Login:        "lead",
			Role:         domain.RoleLead,
			PasswordHash: leadHash,
		},
		"no-password": {
			ID:       "user-3",
			FullName: "No Password",
			Login:    "no-password",
			Role:     domain.RoleIntern,
		},
	})
	if err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}
	return service
}

func TestNewRequiresStorage(t *testing.T) {
	service, err := New(nil)
	if service != nil {
		t.Errorf("New(nil) service = %v, want nil", service)
	}
	if !errors.Is(err, ErrUserStorageRequired) {
		t.Errorf("New(nil) error = %v, want %v", err, ErrUserStorageRequired)
	}
}

func TestAuthenticateReturnsUserWithoutPasswordHash(t *testing.T) {
	service := newTestService(t)

	user, err := service.Authenticate("intern", "intern-password")
	if err != nil {
		t.Fatalf("Authenticate() returned unexpected error: %v", err)
	}
	if user.ID != "user-1" || user.Role != domain.RoleIntern {
		t.Errorf("Authenticate() user = %+v, want intern user", user)
	}
	if user.PasswordHash != "" {
		t.Errorf("Authenticate() exposed password hash %q", user.PasswordHash)
	}
}

func TestAuthenticateRejectsInvalidCredentials(t *testing.T) {
	service := newTestService(t)
	tests := []struct {
		name     string
		login    string
		password string
	}{
		{name: "unknown login", login: "missing", password: "password"},
		{name: "wrong password", login: "intern", password: "wrong"},
		{name: "empty login", login: " ", password: "intern-password"},
		{name: "empty password", login: "intern"},
		{name: "user without password hash", login: "no-password", password: "anything"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.Authenticate(tt.login, tt.password)
			if !errors.Is(err, ErrInvalidCredentials) {
				t.Errorf("Authenticate() error = %v, want %v", err, ErrInvalidCredentials)
			}
			if user != (domain.User{}) {
				t.Errorf("Authenticate() user = %+v, want zero value on failure", user)
			}
		})
	}
}

func TestAuthenticateForChatAllowsInternOnly(t *testing.T) {
	service := newTestService(t)

	user, err := service.AuthenticateForChat("intern", "intern-password")
	if err != nil {
		t.Fatalf("AuthenticateForChat(intern) returned unexpected error: %v", err)
	}
	if user.ID != "user-1" {
		t.Errorf("AuthenticateForChat(intern) user ID = %q, want user-1", user.ID)
	}

	user, err = service.AuthenticateForChat("lead", "lead-password")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("AuthenticateForChat(lead) error = %v, want %v", err, ErrForbidden)
	}
	if user != (domain.User{}) {
		t.Errorf("AuthenticateForChat(lead) user = %+v, want zero value on denial", user)
	}
}

func TestHashPasswordRejectsEmptyPassword(t *testing.T) {
	if _, err := HashPassword(""); err == nil {
		t.Fatal("HashPassword(empty) returned nil error")
	}
}
