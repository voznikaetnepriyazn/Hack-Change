package team

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"unicode"

	parser "github.com/voznikaetnepriyazn/Hack-Change/agent/internal/lib/parse"
)

//go:embed team.md
var description string

//go:embed team.json
var teamJSON []byte

type Employee struct {
	ID               string `json:"id"`
	LastName         string `json:"last_name"`
	FirstName        string `json:"first_name"`
	MiddleName       string `json:"middle_name"`
	FullName         string `json:"full_name"`
	Login            string `json:"login"`
	Password         string `json:"password"`
	PasswordHash     string `json:"password_hash"`
	Role             string `json:"role"`
	Position         string `json:"position"`
	Grade            string `json:"grade"`
	TeamID           string `json:"team_id"`
	Office           string `json:"office"`
	WorkFormat       string `json:"work_format"`
	WorkSchedule     string `json:"work_schedule"`
	GitHubUsername   string `json:"github_username"`
	Email            string `json:"email"`
	Phone            string `json:"phone"`
	Telegram         string `json:"telegram"`
	Status           string `json:"status"`
	StartDate        string `json:"start_date"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	Responsibilities string `json:"responsibilities"`
}

type TeamTool struct{}

func NewTeamTool() *TeamTool {
	return &TeamTool{}
}

func (t *TeamTool) Name() string {
	return "search_team"
}

func (t *TeamTool) Description() string {
	return description
}

func (t *TeamTool) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"query": {
				"type": "string",
				"description": "Запрос стажёра в свободной форме о сотруднике или его роли. Можно передать фразу целиком, например «покажи информацию про менеджера»."
			}
		},
		"required": ["query"],
		"additionalProperties": false
	}`)
}

type searchArgs struct {
	Query string `json:"query"`
}

func (t *TeamTool) Execute(ctx context.Context, argsJSON string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("search team canceled: %w", err)
	}

	args, err := parser.ParseArgs[searchArgs](argsJSON)
	if err != nil {
		slog.Error("Failed to parse team search arguments", "error", err)
		return "", fmt.Errorf("failed to parse team search arguments: %w", err)
	}
	query := strings.TrimSpace(args.Query)
	tokens := searchTokens(query)
	if len(tokens) == 0 {
		return "", errors.New("employee search query is required")
	}

	var employees []Employee
	if err := json.Unmarshal(teamJSON, &employees); err != nil {
		slog.Error("Failed to load team dataset", "error", err)
		return "", fmt.Errorf("failed to load team dataset: %w", err)
	}

	matches := make([]Employee, 0)
	bestScore := 0
	for _, employee := range employees {
		score := employeeMatchScore(employee, tokens)
		if score > bestScore {
			matches = matches[:0]
			bestScore = score
		}
		if score > 0 && score == bestScore {
			matches = append(matches, employee)
		}
	}
	if len(matches) == 0 {
		return fmt.Sprintf("Сотрудник по запросу «%s» не найден.", query), nil
	}

	var result strings.Builder
	for i, employee := range matches {
		if i > 0 {
			result.WriteString("\n\n")
		}
		writeEmployee(&result, employee)
	}
	return result.String(), nil
}

func searchTokens(query string) []string {
	return strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

func employeeMatchScore(employee Employee, tokens []string) int {
	searchable := strings.ToLower(strings.Join([]string{
		employee.FullName,
		employee.Login,
		employee.Role,
		employee.Position,
		employee.Grade,
		employee.TeamID,
		employee.Office,
	}, " "))
	score := 0
	for _, token := range tokens {
		if strings.Contains(searchable, token) || matchesRole(token, employee.Role) {
			score++
		}
	}
	return score
}

var roleSearchTerms = map[string][]string{
	"mentor":   {"ментор", "наставник"},
	"manager":  {"менеджер", "руководитель", "тимлид", "лид"},
	"employee": {"сотрудник", "разработчик", "инженер"},
}

func matchesRole(token, role string) bool {
	for _, term := range roleSearchTerms[role] {
		if editDistanceAtMostOne(normalizeRoleTerm(token), normalizeRoleTerm(term)) {
			return true
		}
	}
	return false
}

func normalizeRoleTerm(term string) string {
	runes := []rune(term)
	for _, suffix := range []string{"ами", "ями", "ов", "ев", "ей", "ом", "ем", "а", "у", "е", "ы", "и", "я", "ю"} {
		suffixRunes := []rune(suffix)
		if len(runes)-len(suffixRunes) >= 4 &&
			strings.HasSuffix(term, suffix) {
			return string(runes[:len(runes)-len(suffixRunes)])
		}
	}
	return term
}

func editDistanceAtMostOne(left, right string) bool {
	leftRunes := []rune(left)
	rightRunes := []rune(right)
	if len(leftRunes)-len(rightRunes) > 1 || len(rightRunes)-len(leftRunes) > 1 {
		return false
	}

	if len(leftRunes) > len(rightRunes) {
		leftRunes, rightRunes = rightRunes, leftRunes
	}

	leftIndex, rightIndex, differences := 0, 0, 0
	for leftIndex < len(leftRunes) && rightIndex < len(rightRunes) {
		if leftRunes[leftIndex] == rightRunes[rightIndex] {
			leftIndex++
			rightIndex++
			continue
		}
		differences++
		if differences > 1 {
			return false
		}
		if len(leftRunes) == len(rightRunes) {
			leftIndex++
		}
		rightIndex++
	}
	if rightIndex < len(rightRunes) || leftIndex < len(leftRunes) {
		differences++
	}
	return differences <= 1
}

func writeEmployee(result *strings.Builder, employee Employee) {
	writeField(result, "Сотрудник", employee.FullName)
	writeField(result, "Логин", employee.Login)
	writeField(result, "Роль", employee.Role)
	writeField(result, "Должность", employee.Position)
	writeField(result, "Грейд", employee.Grade)
	writeField(result, "Команда", employee.TeamID)
	writeField(result, "Офис", employee.Office)
	writeField(result, "Формат работы", employee.WorkFormat)
	writeField(result, "График", employee.WorkSchedule)
	writeField(result, "Статус", employee.Status)
	writeField(result, "Дата начала работы", employee.StartDate)
	writeField(result, "Обязанности", employee.Responsibilities)
}

func writeField(result *strings.Builder, name, value string) {
	if value == "" {
		return
	}
	fmt.Fprintf(result, "%s: %s\n", name, value)
}
