package mergeYaml

import (
	"reflect"
	"testing"
)

func TestMerge(t *testing.T) {
	tests := []struct {
		testcase string
		dst      interface{}
		src      interface{}
		override bool
		expected interface{}
	}{
		// Merging maps
		{
			testcase: "Merging maps",
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
			testcase: "Overriding values in maps",
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
			testcase: "Merging arrays",
			dst:      []interface{}{"value1", "value2"},
			src:      []interface{}{"value3", "value4"},
			override: false,
			expected: []interface{}{"value1", "value2", "value3", "value4"},
		},
		// Overriding values in arrays
		{
			testcase: "Overriding values in arrays",
			dst:      []interface{}{"value1", "value2"},
			src:      []interface{}{"value3", "value4"},
			override: true,
			expected: []interface{}{"value3", "value4"},
		},
		// Extending values with plain values
		{
			testcase: "Extending values with plain values",
			dst:      "value1",
			src:      "value2",
			override: false,
			expected: "value1",
		},
		// Overriding plain values
		{
			testcase: "Overriding plain values",
			dst:      "value1",
			src:      "value2",
			override: true,
			expected: "value2",
		},
	}

	for _, test := range tests {
		result := Merge(test.src, test.dst, test.override)

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("Test case '%s' failed. Expected: %v, Got: %v", test.testcase, test.expected, result)
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

	result := mergeMaps(src, dst, true)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Test failed. Expected: %v, Got: %v", expected, result)
	}
}

func TestMergeLists(t *testing.T) {
	dst := []interface{}{"value1", "value2"}
	src := []interface{}{"value3", "value4"}

	expected := []interface{}{"value1", "value2", "value3", "value4"}

	result := mergeLists(src, dst, false)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Test failed. Expected: %v, Got: %v", expected, result)
	}
}
