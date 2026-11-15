package team

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestTeamToolReturnsStructuredEmployeeCardsWithoutSensitiveData(t *testing.T) {
	result, err := NewTeamTool().Execute(context.Background(), `{"query":"Дмитрий Волков"}`)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}
	var cards []EmployeeCard
	if err := json.Unmarshal([]byte(result), &cards); err != nil {
		t.Fatalf("Execute() returned invalid JSON: %v; result: %s", err, result)
	}
	if len(cards) != 1 {
		t.Fatalf("Execute() returned %d cards, want 1", len(cards))
	}
	card := cards[0]
	if card.FullName != "Волков Дмитрий Сергеевич" {
		t.Errorf("full_name = %q, want Волков Дмитрий Сергеевич", card.FullName)
	}
	if card.Position != "Senior Backend Developer" {
		t.Errorf("position = %q, want Senior Backend Developer", card.Position)
	}
	if card.Grade != "Senior" {
		t.Errorf("grade = %q, want Senior", card.Grade)
	}
	if card.Responsibilities == "" {
		t.Error("responsibilities is empty")
	}
	for _, forbidden := range []string{
		`"id"`,
		`"login"`,
		`"password"`,
		`"password_hash"`,
		`"email"`,
		`"phone"`,
		`"telegram"`,
		`"github_username"`,
		`"created_at"`,
		`"updated_at"`,
		"@avito.ru",
		"@dmitry_volkov",
		"+7",
	} {
		if strings.Contains(strings.ToLower(result), strings.ToLower(forbidden)) {
			t.Errorf("Execute() result contains restricted field %q: %s", forbidden, result)
		}
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
	if strings.TrimSpace(result) != "[]" {
		t.Errorf("Execute() result = %q, want empty JSON array", result)
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
		var cards []EmployeeCard
		if err := json.Unmarshal([]byte(result), &cards); err != nil {
			t.Fatalf("Execute(%q) returned invalid JSON: %v", query, err)
		}
		if len(cards) != 1 || cards[0].FullName != "Морозова Елена Викторовна" {
			t.Errorf("Execute(%q) cards = %+v, want manager card", query, cards)
		}
	}
}
