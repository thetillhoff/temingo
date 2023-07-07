package mergeYaml

import (
	"reflect"
	"testing"
)

func TestMerge(t *testing.T) {
	tests := []struct {
		dst      interface{}
		src      interface{}
		override bool
		expected interface{}
	}{
		// Merging maps
		{
			dst: map[interface{}]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			src: map[interface{}]interface{}{
				"key2": "new_value2",
				"key3": "value3",
			},
			override: false,
			expected: map[interface{}]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		// Overriding values in maps
		{
			dst: map[interface{}]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			src: map[interface{}]interface{}{
				"key2": "new_value2",
				"key3": "value3",
			},
			override: true,
			expected: map[interface{}]interface{}{
				"key1": "value1",
				"key2": "new_value2",
				"key3": "value3",
			},
		},
		// Merging arrays
		{
			dst:      []interface{}{"value1", "value2"},
			src:      []interface{}{"value3", "value4"},
			override: false,
			expected: []interface{}{"value1", "value2", "value3", "value4"},
		},
		// Overriding values in arrays
		{
			dst:      []interface{}{"value1", "value2"},
			src:      []interface{}{"value3", "value4"},
			override: true,
			expected: []interface{}{"value3", "value4"},
		},
		// Extending values with plain values
		{
			dst:      "value1",
			src:      "value2",
			override: false,
			expected: "value2",
		},
		// Overriding plain values
		{
			dst:      "value1",
			src:      "value2",
			override: true,
			expected: "value2",
		},
	}

	for i, test := range tests {
		result := Merge(test.dst, test.src, test.override)

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Test case %d failed. Expected: %v, Got: %v", i+1, test.expected, result)
		}
	}
}

func TestMergeMaps(t *testing.T) {
	dst := map[interface{}]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	src := map[interface{}]interface{}{
		"key2": "new_value2",
		"key3": "value3",
	}

	expected := map[interface{}]interface{}{
		"key1": "value1",
		"key2": "new_value2",
		"key3": "value3",
	}

	result := mergeMaps(dst, src, false)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Test failed. Expected: %v, Got: %v", expected, result)
	}
}

func TestMergeLists(t *testing.T) {
	dst := []interface{}{"value1", "value2"}
	src := []interface{}{"value3", "value4"}

	expected := []interface{}{"value1", "value2", "value3", "value4"}

	result := mergeLists(dst, src, false)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Test failed. Expected: %v, Got: %v", expected, result)
	}
}
