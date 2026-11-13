package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/domain"
)

// Reading loads mock users from a JSON file and keeps them in memory.
// The loaded user set is immutable and safe for concurrent reads.
type Reading struct {
	usersByLogin map[string]domain.User
	usersByID    map[string]domain.User
}

var _ Storage = (*Reading)(nil)

// NewReading loads a JSON array of users from path.
func NewReading(path string) (*Reading, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read users file: %w", err)
	}

	var users []domain.User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("decode users file: %w", err)
	}

	usersByLogin := make(map[string]domain.User, len(users))
	usersByID := make(map[string]domain.User, len(users))
	ids := make(map[string]string, len(users))
	for i, user := range users {
		if user.ID == "" {
			return nil, fmt.Errorf("user at index %d has empty ID", i)
		}
		if user.Login == "" {
			return nil, fmt.Errorf("user at index %d has empty login", i)
		}
		if user.FullName == "" {
			return nil, fmt.Errorf("user %q has empty full name", user.Login)
		}
		if user.Role != domain.RoleIntern && user.Role != domain.RoleLead && user.Role != domain.RoleManager {
			return nil, fmt.Errorf("user %q has invalid role %q", user.Login, user.Role)
		}
		if _, exists := usersByLogin[user.Login]; exists {
			return nil, fmt.Errorf("duplicate user login %q", user.Login)
		}
		if existingLogin, exists := ids[user.ID]; exists {
			return nil, fmt.Errorf("duplicate user ID %q for logins %q and %q", user.ID, existingLogin, user.Login)
		}

		usersByLogin[user.Login] = user
		usersByID[user.ID] = user
		ids[user.ID] = user.Login
	}

	return &Reading{usersByLogin: usersByLogin, usersByID: usersByID}, nil
}

// FindByLogin returns the user with the exact matching login.
func (r *Reading) FindByLogin(login string) (domain.User, bool) {
	user, ok := r.usersByLogin[login]
	return user, ok
}

// FindByID returns the user with the exact matching ID.
func (r *Reading) FindByID(id string) (domain.User, bool) {
	user, ok := r.usersByID[id]
	return user, ok
}
