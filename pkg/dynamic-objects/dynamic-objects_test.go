package dynamic_objects

import (
	"encoding/json"
	"fmt"
	"testing"
)

var jsonString = []byte(`
{
	"sub1": {
		"sub2":{
			"foo":"bar",
			"amount": 1
		}
	}
}
`)

type SubStruct struct {
	Prop string
	Next any
}

func TestGetPropertyByPath(t *testing.T) {
	var o any
	_ = json.Unmarshal(jsonString, &o)
	m, _ := o.(map[string]any)
	s1, _ := m["sub1"].(map[string]any)
	s1["subObj"] = SubStruct{
		Prop: "prop_value",
	}
	tests := []struct {
		path        string
		expected    any
		expectError bool
	}{
		{
			path: "sub1",
			expected: map[string]any{
				"sub2": map[string]any{
					"foo":    "bar",
					"amount": 1,
				},
				"subObj": SubStruct{
					Prop: "prop_value",
				},
			},
		},
		{
			path:     "sub1.sub2.foo",
			expected: "bar",
		},
		{
			path:     "sub1.sub2.amount",
			expected: 1,
		},
		{
			path:     "sub1.subObj.Prop",
			expected: "prop_value",
		},
		{
			path:        "sub1.sub2.does_not_exist",
			expected:    nil,
			expectError: true,
		},
		{
			path:        "sub1.sub2.amount.too_deep",
			expected:    nil,
			expectError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			actual, err := GetPropertyByPath(o, test.path, ".")
			actualJson, _ := json.Marshal(actual)
			expectedJson, _ := json.Marshal(test.expected)
			if string(actualJson) != string(expectedJson) {
				fmt.Printf("Actual: %s\nExpected:%s\n", actualJson, expectedJson)
				t.Fail()
				return
			}
			if test.expectError != (err != nil) {
				t.Fail()
				return
			}
		})
	}
}

func TestSetPropertyByPath(t *testing.T) {
	tests := []struct {
		path     string
		newValue any
	}{
		{
			path:     "sub1",
			newValue: "just a string",
		},
		{
			path:     "sub1.sub2.foo",
			newValue: "other-value",
		},
		{
			path:     "sub1.sub2.amount",
			newValue: 2,
		},
		{
			path:     "sub1.sub2.does_not_exist",
			newValue: "new_property",
		},
		{
			path:     "sub1.sub2.amount.too_deep",
			newValue: "another_new_property",
		},
	}
	for _, test := range tests {
		var o map[string]any
		_ = json.Unmarshal(jsonString, &o)
		t.Run(test.path, func(t *testing.T) {
			SetPropertyByPath(o, test.path, ".", test.newValue, true)
			actual, _ := GetPropertyByPath(o, test.path, ".")
			actualJson, _ := json.Marshal(actual)
			expectedJson, _ := json.Marshal(test.newValue)
			if string(actualJson) != string(expectedJson) {
				fmt.Printf("Actual: %s\nExpected:%s\n", actualJson, expectedJson)
				t.Fail()
				return
			}
		})
	}
}

func TestPivotStruct(t *testing.T) {
	var nilPtr *SubStruct
	nonNilPtr := &SubStruct{
		Prop: "Hello",
		Next: nil,
	}
	o := SubStruct{
		Prop: "",
		Next: SubStruct{
			Prop: "value",
			Next: map[string]any{
				"struct_in_map": SubStruct{
					Prop: "deep_value",
					Next: nilPtr,
				},
				"struct_in_map_with_non_nil_ptr": SubStruct{
					Prop: "",
					Next: nonNilPtr,
				},
			},
		},
	}
	for i := 0; i < 10000; i++ {
		bytes, _ := json.Marshal(PivotAndJoinPaths(o, "/"))
		s := string(bytes)
		fmt.Println()
		if s != `["Prop","Next/Prop","Next/Next/struct_in_map/Prop","Next/Next/struct_in_map/Next","Next/Next/struct_in_map_with_non_nil_ptr/Prop","Next/Next/struct_in_map_with_non_nil_ptr/Next/Prop","Next/Next/struct_in_map_with_non_nil_ptr/Next/Next"]` &&
			s != `["Prop","Next/Prop","Next/Next/struct_in_map_with_non_nil_ptr/Prop","Next/Next/struct_in_map_with_non_nil_ptr/Next/Prop","Next/Next/struct_in_map_with_non_nil_ptr/Next/Next","Next/Next/struct_in_map/Prop","Next/Next/struct_in_map/Next"]` {
			t.FailNow()
		}
	}
}
