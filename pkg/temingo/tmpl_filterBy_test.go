package temingo

import (
	"reflect"
	"testing"
)

func TestTmplFilterBy(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    any
		input    map[string]any
		expected map[string]any
	}{
		{
			name:     "Nil map",
			field:    "publish",
			value:    true,
			input:    nil,
			expected: map[string]any{},
		},
		{
			name:     "Empty map",
			field:    "publish",
			value:    true,
			input:    map[string]any{},
			expected: map[string]any{},
		},
		{
			name:  "No field — keep (default-include)",
			field: "publish",
			value: true,
			input: map[string]any{
				"a": map[string]any{"title": "A"},
			},
			expected: map[string]any{
				"a": map[string]any{"title": "A"},
			},
		},
		{
			name:  "Field matches — keep",
			field: "publish",
			value: true,
			input: map[string]any{
				"a": map[string]any{"title": "A", "publish": true},
			},
			expected: map[string]any{
				"a": map[string]any{"title": "A", "publish": true},
			},
		},
		{
			name:  "Field does not match — drop",
			field: "publish",
			value: true,
			input: map[string]any{
				"a": map[string]any{"title": "A", "publish": false},
			},
			expected: map[string]any{},
		},
		{
			name:  "Mixed — keep absent and matching, drop non-matching",
			field: "publish",
			value: true,
			input: map[string]any{
				"a": map[string]any{"title": "A"},
				"b": map[string]any{"title": "B", "publish": true},
				"c": map[string]any{"title": "C", "publish": false},
			},
			expected: map[string]any{
				"a": map[string]any{"title": "A"},
				"b": map[string]any{"title": "B", "publish": true},
			},
		},
		{
			name:  "String field match",
			field: "status",
			value: "active",
			input: map[string]any{
				"a": map[string]any{"status": "active"},
				"b": map[string]any{"status": "draft"},
				"c": map[string]any{"title": "no status"},
			},
			expected: map[string]any{
				"a": map[string]any{"status": "active"},
				"c": map[string]any{"title": "no status"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tmpl_filterBy(tt.field, tt.value, tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("tmpl_filterBy(%q, %v) =\n  %v\nwant\n  %v", tt.field, tt.value, result, tt.expected)
			}
		})
	}
}
