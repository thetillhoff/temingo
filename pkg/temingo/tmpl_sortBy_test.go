package temingo

import (
	"reflect"
	"testing"
)

func TestTmplSortBy(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		input    map[string]any
		expected []any
	}{
		{
			name:     "Nil map",
			field:    "date",
			input:    nil,
			expected: []any{},
		},
		{
			name:     "Empty map",
			field:    "date",
			input:    map[string]any{},
			expected: []any{},
		},
		{
			name:  "Single entry",
			field: "date",
			input: map[string]any{
				"post-a": map[string]any{"date": "2026-01-01", "title": "A"},
			},
			expected: []any{
				map[string]any{"key": "post-a", "value": map[string]any{"date": "2026-01-01", "title": "A"}},
			},
		},
		{
			name:  "Sort by date ascending",
			field: "date",
			input: map[string]any{
				"post-b": map[string]any{"date": "2026-03-01", "title": "B"},
				"post-a": map[string]any{"date": "2026-01-01", "title": "A"},
				"post-c": map[string]any{"date": "2026-02-01", "title": "C"},
			},
			expected: []any{
				map[string]any{"key": "post-a", "value": map[string]any{"date": "2026-01-01", "title": "A"}},
				map[string]any{"key": "post-c", "value": map[string]any{"date": "2026-02-01", "title": "C"}},
				map[string]any{"key": "post-b", "value": map[string]any{"date": "2026-03-01", "title": "B"}},
			},
		},
		{
			name:  "Missing field sorts last, tie-broken by key",
			field: "date",
			input: map[string]any{
				"post-b": map[string]any{"date": "2026-01-01", "title": "B"},
				"post-a": map[string]any{"title": "A"},
				"post-c": map[string]any{"title": "C"},
			},
			expected: []any{
				map[string]any{"key": "post-b", "value": map[string]any{"date": "2026-01-01", "title": "B"}},
				map[string]any{"key": "post-a", "value": map[string]any{"title": "A"}},
				map[string]any{"key": "post-c", "value": map[string]any{"title": "C"}},
			},
		},
		{
			name:  "Sort by title",
			field: "title",
			input: map[string]any{
				"post-z": map[string]any{"title": "Zebra"},
				"post-a": map[string]any{"title": "Apple"},
				"post-m": map[string]any{"title": "Mango"},
			},
			expected: []any{
				map[string]any{"key": "post-a", "value": map[string]any{"title": "Apple"}},
				map[string]any{"key": "post-m", "value": map[string]any{"title": "Mango"}},
				map[string]any{"key": "post-z", "value": map[string]any{"title": "Zebra"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tmpl_sortBy(tt.field, tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("tmpl_sortBy(%q, ...) =\n  %v\nwant\n  %v", tt.field, result, tt.expected)
			}
		})
	}
}
