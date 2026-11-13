package team

import (
	"context"
	"strings"
	"testing"
)

func TestTeamToolReturnsTextWithoutCredentialsOrContacts(t *testing.T) {
	result, err := NewTeamTool().Execute(context.Background(), `{"query":"Дмитрий Волков"}`)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	for _, want := range []string{
		"Сотрудник: Волков Дмитрий Сергеевич",
		"Должность: Senior Backend Developer",
		"Грейд: Senior",
		"Обязанности:",
	} {
		if !strings.Contains(result, want) {
			t.Errorf("Execute() result does not contain %q: %s", want, result)
		}
	}
	for _, forbidden := range []string{"password", "password_hash", "@avito.ru", "@dmitry_volkov", "+7"} {
		if strings.Contains(strings.ToLower(result), strings.ToLower(forbidden)) {
			t.Errorf("Execute() result contains restricted field %q: %s", forbidden, result)
		}
	}
	if strings.HasPrefix(strings.TrimSpace(result), "{") {
		t.Errorf("Execute() should return text, got JSON: %s", result)
	}
}

func TestTeamToolRejectsEmptyQuery(t *testing.T) {
	_, err := NewTeamTool().Execute(context.Background(), `{"query":"  "}`)
	if err == nil {
		t.Fatal("Execute() with an empty query returned no error")
	}
}

func TestTeamToolReportsNoMatches(t *testing.T) {
	result, err := NewTeamTool().Execute(context.Background(), `{"query":"unknown person"}`)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	if !strings.Contains(result, "не найден") {
		t.Errorf("Execute() result = %q, want a not-found message", result)
	}
}

func TestTeamToolFindsManagerFromNaturalLanguageRequest(t *testing.T) {
	for _, query := range []string{
		"выведи мне инфу про менеджера",
		"веведи мне инфу про тенеджера",
	} {
		result, err := NewTeamTool().Execute(context.Background(), `{"query":"`+query+`"}`)
		if err != nil {
			t.Fatalf("Execute(%q) returned error: %v", query, err)
		}
		if !strings.Contains(result, "Сотрудник: Морозова Елена Викторовна") {
			t.Errorf("Execute(%q) result = %q, want manager profile", query, result)
		}
	}
}
