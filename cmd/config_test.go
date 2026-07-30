package cmd

import (
	"reflect"
	"testing"

	"github.com/thetillhoff/temingo/pkg/refcheck"
)

func TestAllowlistFromConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		expected refcheck.Allowlist
	}{
		{
			name:     "absent key yields nothing",
			config:   map[string]interface{}{},
			expected: nil,
		},
		{
			name: "entry without checks",
			config: map[string]interface{}{
				"allow": []interface{}{
					map[string]interface{}{"url": "https://paywalled.example/*"},
				},
			},
			expected: refcheck.Allowlist{{URL: "https://paywalled.example/*"}},
		},
		{
			name: "entry with checks",
			config: map[string]interface{}{
				"allow": []interface{}{
					map[string]interface{}{
						"url":    "https://redirecting.example/*",
						"checks": []interface{}{"redirect", "status"},
					},
				},
			},
			expected: refcheck.Allowlist{{
				URL:    "https://redirecting.example/*",
				Checks: []refcheck.Category{refcheck.CategoryRedirect, refcheck.CategoryStatus},
			}},
		},
		{
			name: "malformed entries are skipped",
			config: map[string]interface{}{
				"allow": []interface{}{"not-a-map", map[string]interface{}{"nourl": "x"}},
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := allowlistFromConfig(test.config)
			if !reflect.DeepEqual(got, test.expected) {
				t.Errorf("allowlistFromConfig() = %+v, want %+v", got, test.expected)
			}
		})
	}
}
