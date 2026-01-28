package pipeline

import (
	"reflect"
	"testing"
)

func TestDifference_Ints(t *testing.T) {
	tests := []struct {
		name   string
		first  []int
		second []int
		want   []int
	}{
		{name: "both nil -> empty", first: nil, second: nil, want: []int{}},
		{name: "first nil -> empty", first: nil, second: []int{1, 2}, want: []int{}},
		{name: "second nil -> empty", first: []int{1, 2}, second: nil, want: []int{}},
		{name: "both empty -> empty", first: []int{}, second: []int{}, want: []int{}},
		{name: "second empty (non-nil) -> copy of first", first: []int{1, 2, 3}, second: []int{}, want: []int{1, 2, 3}},
		{name: "no overlap -> all from first", first: []int{1, 2, 3}, second: []int{4, 5}, want: []int{1, 2, 3}},
		{name: "partial overlap -> only uniques from first", first: []int{1, 2, 3}, second: []int{2, 4}, want: []int{1, 3}},
		{name: "all overlap -> empty", first: []int{1, 2}, second: []int{2, 1}, want: []int{}},
		{name: "duplicates in first preserved if not in second", first: []int{1, 1, 2, 3, 3, 3}, second: []int{2}, want: []int{1, 1, 3, 3, 3}},
		{name: "duplicates in first removed if present in second", first: []int{1, 1, 2, 2}, second: []int{1, 2}, want: []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Difference(tt.first, tt.second)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Difference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDifference_Strings(t *testing.T) {
	first := []string{"apple", "banana", "cherry", "banana"}
	second := []string{"banana", "date"}
	want := []string{"apple", "cherry"}

	got := Difference(first, second)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Difference() = %v, want %v", got, want)
	}
}
