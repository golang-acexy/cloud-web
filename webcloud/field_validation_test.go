package webcloud

import (
	"errors"
	"reflect"
	"testing"
)

type fieldValidationDTO struct {
	ID           int
	UserID       int
	Name         string
	CreatedAt    string
	CreateTime   string
	ModifiedAt   string
	ModifiedTime string
	UpdateTime   string
	UpdateAt     string
	UpdatedAt    string
	DeletedAt    string
	Hidden       string `json:"-"`
}

// TestFieldConversionContract preserves key rules, values, and input ownership.
func TestFieldConversionContract(t *testing.T) {
	router := NewBaseRouter[int, struct{}, fieldValidationDTO, fieldValidationDTO, struct{}](nil)
	for _, tc := range []struct {
		name    string
		input   map[string]any
		mode    mode
		want    map[string]any
		invalid bool
	}{
		{name: "camel case and zero", input: map[string]any{"userId": 0, "name": ""}, want: map[string]any{"user_id": 0, "name": ""}},
		{name: "snake case and null", input: map[string]any{"user_id": nil, "name": false}, want: map[string]any{"user_id": nil, "name": false}},
		{name: "query ID", input: map[string]any{"id": 0}, want: map[string]any{"id": 0}},
		{name: "JSON tags remain ignored", input: map[string]any{"hidden": nil}, want: map[string]any{"hidden": nil}},
		{name: "soft deletion remains allowed", mode: modifyMode, input: map[string]any{"deletedAt": nil}, want: map[string]any{"deleted_at": nil}},
		{name: "unknown field", input: map[string]any{"name": "ok", "unknownField": 1}, invalid: true},
		{name: "unknown modification", mode: modifyMode, input: map[string]any{"name": "ok", "unknownField": 1}, invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			original := make(map[string]any, len(tc.input))
			for k, v := range tc.input {
				original[k] = v
			}
			got, err := router.convertFieldsToColumns(tc.input, tc.mode)
			if tc.invalid {
				if !errors.Is(err, ErrBadRequestParameters) || got != nil {
					t.Fatalf("unexpected rejection: %v %v", got, err)
				}
			} else {
				if err != nil || !reflect.DeepEqual(got, tc.want) {
					t.Fatalf("got %v %v, want %v", got, err, tc.want)
				}
				got["output_only"] = true
			}
			if !reflect.DeepEqual(tc.input, original) {
				t.Fatal("input map was mutated")
			}
		})
	}
	for _, column := range []string{"id", "created_at", "create_time", "modified_at", "modified_time", "update_time", "update_at", "updated_at"} {
		if _, err := router.convertFieldsToColumns(map[string]any{column: 0}, queryMode); err != nil {
			t.Fatalf("query rejected %s", column)
		}
		if got, err := router.convertFieldsToColumns(map[string]any{column: 0}, modifyMode); !errors.Is(err, ErrBadRequestParameters) || got != nil {
			t.Fatalf("modification accepted %s", column)
		}
	}
	// Existing alias collisions remain accepted, with no new winner-order guarantee.
	for i := 0; i < 100; i++ {
		got, err := router.convertFieldsToColumns(map[string]any{"userId": 1, "user_id": 2}, queryMode)
		if err != nil || len(got) != 1 || (got["user_id"] != 1 && got["user_id"] != 2) {
			t.Fatalf("alias semantics changed: %v %v", got, err)
		}
	}
}
