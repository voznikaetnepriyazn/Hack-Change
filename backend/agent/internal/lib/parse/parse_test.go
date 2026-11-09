package parser

import (
	"strings"
	"testing"
)

func TestParseArgsStruct(t *testing.T) {
	type arguments struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}

	got, err := ParseArgs[arguments](`{"query":"weather","limit":5}`)
	if err != nil {
		t.Fatalf("ParseArgs() returned unexpected error: %v", err)
	}

	want := arguments{Query: "weather", Limit: 5}
	if got != want {
		t.Errorf("ParseArgs() = %+v, want %+v", got, want)
	}
}

func TestParseArgsMap(t *testing.T) {
	got, err := ParseArgs[map[string]int](`{"first":1,"second":2}`)
	if err != nil {
		t.Fatalf("ParseArgs() returned unexpected error: %v", err)
	}

	want := map[string]int{"first": 1, "second": 2}
	if len(got) != len(want) {
		t.Fatalf("ParseArgs() returned %d entries, want %d", len(got), len(want))
	}
	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Errorf("ParseArgs()[%q] = %d, want %d", key, got[key], wantValue)
		}
	}
}

func TestParseArgsJSONValues(t *testing.T) {
	tests := []struct {
		name string
		json string
		want []string
	}{
		{
			name: "array",
			json: `["a","b"]`,
			want: []string{"a", "b"},
		},
		{
			name: "string",
			json: `"hello"`,
			want: []string{"hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "array":
				got, err := ParseArgs[[]string](tt.json)
				if err != nil {
					t.Fatalf("ParseArgs() returned unexpected error: %v", err)
				}
				if len(got) != len(tt.want) {
					t.Fatalf("ParseArgs() returned %d values, want %d", len(got), len(tt.want))
				}
				for i, want := range tt.want {
					if got[i] != want {
						t.Errorf("ParseArgs()[%d] = %q, want %q", i, got[i], want)
					}
				}
			case "string":
				got, err := ParseArgs[string](tt.json)
				if err != nil {
					t.Fatalf("ParseArgs() returned unexpected error: %v", err)
				}
				if got != tt.want[0] {
					t.Errorf("ParseArgs() = %q, want %q", got, tt.want[0])
				}
			}
		})
	}
}

func TestParseArgsInvalidJSON(t *testing.T) {
	got, err := ParseArgs[map[string]string](`{"query":`)
	if err == nil {
		t.Fatal("ParseArgs() returned nil error for malformed JSON")
	}
	if !strings.HasPrefix(err.Error(), "failed to parse arguments:") {
		t.Errorf("ParseArgs() error = %q, want parse error context", err)
	}
	if got != nil {
		t.Errorf("ParseArgs() result = %#v, want nil zero value", got)
	}
}

func TestParseArgsIncompatibleType(t *testing.T) {
	type arguments struct {
		Count int `json:"count"`
	}

	got, err := ParseArgs[arguments](`{"count":"not a number"}`)
	if err == nil {
		t.Fatal("ParseArgs() returned nil error for incompatible JSON value")
	}
	if !strings.HasPrefix(err.Error(), "failed to parse arguments:") {
		t.Errorf("ParseArgs() error = %q, want parse error context", err)
	}
	if got != (arguments{}) {
		t.Errorf("ParseArgs() result = %+v, want zero value", got)
	}
}

func TestParseArgsNull(t *testing.T) {
	got, err := ParseArgs[*struct{ Name string }](`null`)
	if err != nil {
		t.Fatalf("ParseArgs() returned unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("ParseArgs() = %#v, want nil", got)
	}
}
