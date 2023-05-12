package helper

import (
	"miniK8s/pkg/apiObject"
	"reflect"
	"testing"
)

func TestGetEndpoints(t *testing.T) {
	// 创建测试用例
	tests := []struct {
		name     string
		key      string
		value    string
		expected []apiObject.Endpoint
	}{
		{
			name:     "Test Case 1",
			key:      "key1",
			value:    "value1",
			expected: []apiObject.Endpoint{},
		},
		{
			name:     "Test Case 2",
			key:      "key2",
			value:    "value2",
			expected: []apiObject.Endpoint{},
		},
		// 可以添加更多的测试用例
	}

	// 执行测试
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			endpoints, err := GetEndpoints(test.key, test.value)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(endpoints, test.expected) {
				t.Errorf("expected %+v, but got %+v", test.expected, endpoints)
			}
		})
	}
}
