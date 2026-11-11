package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/voznikaetnepriyazn/Hack-Change/agent/internal/auth/domain"
)

func writeUsersFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "users.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write users file: %v", err)
	}
	return path
}

func TestNewReadingLoadsUsersAndFindsByLogin(t *testing.T) {
	path := writeUsersFile(t, `[
		{"id":"u-1","full_name":"Alex Intern","login":"alex","role":"intern"},
		{"id":"u-2","full_name":"Sam Lead","login":"sam","role":"lead"}
	]`)

	users, err := NewReading(path)
	if err != nil {
		t.Fatalf("NewReading() returned unexpected error: %v", err)
	}

	tests := []struct {
		login string
		want  domain.User
		ok    bool
	}{
		{
			login: "alex",
			want:  domain.User{ID: "u-1", FullName: "Alex Intern", Login: "alex", Role: domain.RoleIntern},
			ok:    true,
		},
		{
			login: "sam",
			want:  domain.User{ID: "u-2", FullName: "Sam Lead", Login: "sam", Role: domain.RoleLead},
			ok:    true,
		},
		{login: "missing", ok: false},
	}

	for _, tt := range tests {
		got, ok := users.FindByLogin(tt.login)
		if ok != tt.ok {
			t.Errorf("FindByLogin(%q) found = %t, want %t", tt.login, ok, tt.ok)
		}
		if ok && got != tt.want {
			t.Errorf("FindByLogin(%q) = %+v, want %+v", tt.login, got, tt.want)
		}
	}
}

func TestNewReadingReportsInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		json string
		want string
	}{
		{
			name: "invalid JSON",
			json: `[`,
			want: "decode users file:",
		},
		{
			name: "empty ID",
			json: `[{"full_name":"Alex","login":"alex","role":"intern"}]`,
			want: "empty ID",
		},
		{
			name: "empty login",
			json: `[{"id":"u-1","full_name":"Alex","role":"intern"}]`,
			want: "empty login",
		},
		{
			name: "empty full name",
			json: `[{"id":"u-1","login":"alex","role":"intern"}]`,
			want: "empty full name",
		},
		{
			name: "unknown role",
			json: `[{"id":"u-1","full_name":"Alex","login":"alex","role":"admin"}]`,
			want: "invalid role",
		},
		{
			name: "duplicate login",
			json: `[
				{"id":"u-1","full_name":"Alex","login":"alex","role":"intern"},
				{"id":"u-2","full_name":"Other Alex","login":"alex","role":"lead"}
			]`,
			want: "duplicate user login",
		},
		{
			name: "duplicate ID",
			json: `[
				{"id":"u-1","full_name":"Alex","login":"alex","role":"intern"},
				{"id":"u-1","full_name":"Sam","login":"sam","role":"lead"}
			]`,
			want: "duplicate user ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewReading(writeUsersFile(t, tt.json))
			if err == nil {
				t.Fatal("NewReading() returned nil error for invalid input")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("NewReading() error = %q, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestNewReadingReportsMissingFile(t *testing.T) {
	_, err := NewReading(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("NewReading() returned nil error for a missing file")
	}
	if !strings.Contains(err.Error(), "read users file:") {
		t.Errorf("NewReading() error = %q, want read error context", err)
	}
}

func TestNewReadingStoresPasswordHashFromMockFile(t *testing.T) {
	users, err := NewReading(writeUsersFile(t, `[
		{"id":"u-1","full_name":"Alex","login":"alex","role":"intern","password_hash":"mock-value"}
	]`))
	if err != nil {
		t.Fatalf("NewReading() returned unexpected error: %v", err)
	}

	user, ok := users.FindByLogin("alex")
	if !ok {
		t.Fatal("FindByLogin() did not find loaded user")
	}
	if user.PasswordHash != "mock-value" {
		t.Errorf("loaded password hash = %q, want mock value", user.PasswordHash)
	}
}
