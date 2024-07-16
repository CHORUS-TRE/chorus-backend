package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_RemoveDuplicates_uint64(t *testing.T) {

	tests := []struct {
		Name         string
		List         []uint64
		ExpectedList []uint64
	}{
		{
			Name:         "Empty list",
			List:         []uint64{},
			ExpectedList: []uint64{},
		},
		{
			Name:         "No duplicates uint64",
			List:         []uint64{1, 2, 3, 4, 5},
			ExpectedList: []uint64{1, 2, 3, 4, 5},
		},
		{
			Name:         "Duplicates",
			List:         []uint64{1, 2, 2, 3, 4, 4, 5},
			ExpectedList: []uint64{1, 2, 3, 4, 5},
		},
		{
			Name:         "Unordered duplicates",
			List:         []uint64{1, 2, 3, 2, 4, 3, 4, 5},
			ExpectedList: []uint64{1, 2, 3, 4, 5},
		},
		{
			Name:         "More than 2 duplicates",
			List:         []uint64{1, 2, 1, 2, 4, 2, 5, 2, 4, 3, 4, 1, 4, 5},
			ExpectedList: []uint64{1, 2, 3, 4, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			res := RemoveDuplicates(tt.List)

			require.ElementsMatch(t, res, tt.ExpectedList)
		})
	}
}

func Test_RemoveDuplicates_string(t *testing.T) {

	tests := []struct {
		Name         string
		List         []string
		ExpectedList []string
	}{
		{
			Name:         "Empty list",
			List:         []string{},
			ExpectedList: []string{},
		},
		{
			Name:         "No duplicates uint64",
			List:         []string{"1", "def", "3", "abc", "5"},
			ExpectedList: []string{"1", "def", "3", "abc", "5"},
		},
		{
			Name:         "Duplicates",
			List:         []string{"1", "def", "def", "3", "abc", "abc", "5"},
			ExpectedList: []string{"1", "def", "3", "abc", "5"},
		},
		{
			Name:         "Unordered duplicates",
			List:         []string{"1", "def", "3", "def", "abc", "3", "abc", "5"},
			ExpectedList: []string{"1", "def", "3", "abc", "5"},
		},
		{
			Name:         "More than 2 duplicates",
			List:         []string{"1", "def", "1", "def", "abc", "def", "5", "def", "abc", "3", "abc", "1", "abc", "5", "abc", "def", "abc"},
			ExpectedList: []string{"1", "def", "3", "abc", "5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			res := RemoveDuplicates(tt.List)

			require.ElementsMatch(t, res, tt.ExpectedList)
		})
	}
}

func Test_RemoveCommonElementsInLists_uint64(t *testing.T) {

	tests := []struct {
		Name          string
		ListA         []uint64
		ListB         []uint64
		ExpectedListA []uint64
		ExpectedListB []uint64
	}{
		{
			Name:          "Empty lists",
			ListA:         []uint64{},
			ListB:         []uint64{},
			ExpectedListA: []uint64{},
			ExpectedListB: []uint64{},
		},
		{
			Name:          "First list empty",
			ListA:         []uint64{},
			ListB:         []uint64{4, 5},
			ExpectedListA: []uint64{},
			ExpectedListB: []uint64{4, 5},
		},
		{
			Name:          "Second list empty",
			ListA:         []uint64{1, 2},
			ListB:         []uint64{},
			ExpectedListA: []uint64{1, 2},
			ExpectedListB: []uint64{},
		},
		{
			Name:          "No common elements",
			ListA:         []uint64{1, 2},
			ListB:         []uint64{4, 5},
			ExpectedListA: []uint64{1, 2},
			ExpectedListB: []uint64{4, 5},
		},
		{
			Name:          "One common element",
			ListA:         []uint64{1, 2, 3},
			ListB:         []uint64{3, 4, 5},
			ExpectedListA: []uint64{1, 2},
			ExpectedListB: []uint64{4, 5},
		},
		{
			Name:          "Multiple common elements",
			ListA:         []uint64{1, 2, 3, 6},
			ListB:         []uint64{3, 4, 5, 6},
			ExpectedListA: []uint64{1, 2},
			ExpectedListB: []uint64{4, 5},
		},
		{
			Name:          "Duplicates in list A",
			ListA:         []uint64{1, 2, 3, 3},
			ListB:         []uint64{},
			ExpectedListA: []uint64{1, 2, 3, 3},
			ExpectedListB: []uint64{},
		},
		{
			Name:          "Duplicates in list B",
			ListA:         []uint64{},
			ListB:         []uint64{4, 5, 6, 6},
			ExpectedListA: []uint64{},
			ExpectedListB: []uint64{4, 5, 6, 6},
		},
		{
			Name:          "Duplicates and common elements",
			ListA:         []uint64{1, 2, 3, 3, 4},
			ListB:         []uint64{4, 5, 6, 6},
			ExpectedListA: []uint64{1, 2, 3, 3},
			ExpectedListB: []uint64{5, 6, 6},
		},
		{
			Name:          "Same duplicates in lists",
			ListA:         []uint64{1, 2, 3, 3, 4},
			ListB:         []uint64{4, 5, 3, 3, 6, 6},
			ExpectedListA: []uint64{1, 2},
			ExpectedListB: []uint64{5, 6, 6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			resA, resB := RemoveCommonElementsInLists(tt.ListA, tt.ListB)

			require.ElementsMatch(t, resA, tt.ExpectedListA)
			require.ElementsMatch(t, resB, tt.ExpectedListB)
		})
	}
}

func Test_RemoveCommonElementsInLists_string(t *testing.T) {

	tests := []struct {
		Name          string
		ListA         []string
		ListB         []string
		ExpectedListA []string
		ExpectedListB []string
	}{
		{
			Name:          "Empty lists",
			ListA:         []string{},
			ListB:         []string{},
			ExpectedListA: []string{},
			ExpectedListB: []string{},
		},
		{
			Name:          "First list empty",
			ListA:         []string{},
			ListB:         []string{"abc", "5"},
			ExpectedListA: []string{},
			ExpectedListB: []string{"abc", "5"},
		},
		{
			Name:          "Second list empty",
			ListA:         []string{"1", "def"},
			ListB:         []string{},
			ExpectedListA: []string{"1", "def"},
			ExpectedListB: []string{},
		},
		{
			Name:          "No common elements",
			ListA:         []string{"1", "def"},
			ListB:         []string{"abc", "5"},
			ExpectedListA: []string{"1", "def"},
			ExpectedListB: []string{"abc", "5"},
		},
		{
			Name:          "One common element",
			ListA:         []string{"1", "def", "3"},
			ListB:         []string{"3", "abc", "5"},
			ExpectedListA: []string{"1", "def"},
			ExpectedListB: []string{"abc", "5"},
		},
		{
			Name:          "Multiple common elements",
			ListA:         []string{"1", "def", "3", "6"},
			ListB:         []string{"3", "abc", "5", "6"},
			ExpectedListA: []string{"1", "def"},
			ExpectedListB: []string{"abc", "5"},
		},
		{
			Name:          "Duplicates in list A",
			ListA:         []string{"1", "def", "3", "3"},
			ListB:         []string{},
			ExpectedListA: []string{"1", "def", "3", "3"},
			ExpectedListB: []string{},
		},
		{
			Name:          "Duplicates in list B",
			ListA:         []string{},
			ListB:         []string{"abc", "5", "6", "6"},
			ExpectedListA: []string{},
			ExpectedListB: []string{"abc", "5", "6", "6"},
		},
		{
			Name:          "Duplicates and common elements",
			ListA:         []string{"1", "def", "3", "3", "abc"},
			ListB:         []string{"abc", "5", "6", "6"},
			ExpectedListA: []string{"1", "def", "3", "3"},
			ExpectedListB: []string{"5", "6", "6"},
		},
		{
			Name:          "Same duplicates in lists",
			ListA:         []string{"1", "def", "3", "3", "abc"},
			ListB:         []string{"abc", "5", "3", "3", "6", "6"},
			ExpectedListA: []string{"1", "def"},
			ExpectedListB: []string{"5", "6", "6"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			resA, resB := RemoveCommonElementsInLists(tt.ListA, tt.ListB)

			require.ElementsMatch(t, resA, tt.ExpectedListA)
			require.ElementsMatch(t, resB, tt.ExpectedListB)
		})
	}
}

func Test_AreUnorderedListsEqual_uint64(t *testing.T) {

	tests := []struct {
		Name           string
		ListA          []uint64
		ListB          []uint64
		ExpectedResult bool
	}{
		{
			Name:           "Empty lists",
			ListA:          []uint64{},
			ListB:          []uint64{},
			ExpectedResult: true,
		},
		{
			Name:           "First list empty",
			ListA:          []uint64{},
			ListB:          []uint64{4, 5},
			ExpectedResult: false,
		},
		{
			Name:           "Second list empty",
			ListA:          []uint64{1, 2},
			ListB:          []uint64{},
			ExpectedResult: false,
		},
		{
			Name:           "No common elements",
			ListA:          []uint64{1, 2},
			ListB:          []uint64{4, 5},
			ExpectedResult: false,
		},
		{
			Name:           "One common element",
			ListA:          []uint64{1, 2, 3},
			ListB:          []uint64{3, 4, 5},
			ExpectedResult: false,
		},
		{
			Name:           "Multiple common elements",
			ListA:          []uint64{1, 2, 3, 6},
			ListB:          []uint64{3, 4, 5, 6},
			ExpectedResult: false,
		},
		{
			Name:           "Equal lists",
			ListA:          []uint64{1, 2, 3, 6},
			ListB:          []uint64{1, 2, 3, 6},
			ExpectedResult: true,
		},
		{
			Name:           "Unordered equal lists",
			ListA:          []uint64{1, 2, 3, 6},
			ListB:          []uint64{2, 6, 3, 1},
			ExpectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			res := AreUnorderedListsEqual(tt.ListA, tt.ListB)
			require.Equal(t, res, tt.ExpectedResult)
		})
	}
}

func Test_AreUnorderedListsEqual_string(t *testing.T) {

	tests := []struct {
		Name           string
		ListA          []string
		ListB          []string
		ExpectedResult bool
	}{
		{
			Name:           "Empty lists",
			ListA:          []string{},
			ListB:          []string{},
			ExpectedResult: true,
		},
		{
			Name:           "First list empty",
			ListA:          []string{},
			ListB:          []string{"abc", "5"},
			ExpectedResult: false,
		},
		{
			Name:           "Second list empty",
			ListA:          []string{"1", "def"},
			ListB:          []string{},
			ExpectedResult: false,
		},
		{
			Name:           "No common elements",
			ListA:          []string{"1", "def"},
			ListB:          []string{"abc", "5"},
			ExpectedResult: false,
		},
		{
			Name:           "One common element",
			ListA:          []string{"1", "def", "3"},
			ListB:          []string{"3", "abc", "5"},
			ExpectedResult: false,
		},
		{
			Name:           "Multiple common elements",
			ListA:          []string{"1", "def", "3", "6"},
			ListB:          []string{"3", "abc", "5", "6"},
			ExpectedResult: false,
		},
		{
			Name:           "Equal lists",
			ListA:          []string{"1", "def", "3", "6"},
			ListB:          []string{"1", "def", "3", "6"},
			ExpectedResult: true,
		},
		{
			Name:           "Unordered equal lists",
			ListA:          []string{"1", "def", "3", "6"},
			ListB:          []string{"def", "6", "3", "1"},
			ExpectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			res := AreUnorderedListsEqual(tt.ListA, tt.ListB)
			require.Equal(t, res, tt.ExpectedResult)
		})
	}
}

func Test_GetElementsNotInList_uint64(t *testing.T) {

	tests := []struct {
		Name             string
		List             []uint64
		Elements         []uint64
		ExpectedElements []uint64
	}{
		{
			Name:             "Empty list and elements",
			List:             []uint64{},
			Elements:         []uint64{},
			ExpectedElements: []uint64{},
		},
		{
			Name:             "Empty list",
			List:             []uint64{},
			Elements:         []uint64{4, 5},
			ExpectedElements: []uint64{4, 5},
		},
		{
			Name:             "Empty elements",
			List:             []uint64{1, 2},
			Elements:         []uint64{},
			ExpectedElements: []uint64{},
		},
		{
			Name:             "Elements present in list",
			List:             []uint64{1, 2, 3, 4},
			Elements:         []uint64{3, 4, 5},
			ExpectedElements: []uint64{5},
		},
		{
			Name:             "Elements present in unordered list",
			List:             []uint64{4, 2, 1, 3},
			Elements:         []uint64{3, 4, 5},
			ExpectedElements: []uint64{5},
		},
		{
			Name:             "Unordered elements present in list",
			List:             []uint64{1, 2, 3, 4},
			Elements:         []uint64{5, 3, 4},
			ExpectedElements: []uint64{5},
		},
		{
			Name:             "Unordered elements present in unordered list",
			List:             []uint64{4, 2, 1, 3},
			Elements:         []uint64{5, 3, 4},
			ExpectedElements: []uint64{5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			res := GetElementsNotInList(tt.List, tt.Elements)

			require.ElementsMatch(t, res, tt.ExpectedElements)
		})
	}
}

func Test_GetElementsNotInList_string(t *testing.T) {

	tests := []struct {
		Name             string
		List             []string
		Elements         []string
		ExpectedElements []string
	}{
		{
			Name:             "Empty list and elements",
			List:             []string{},
			Elements:         []string{},
			ExpectedElements: []string{},
		},
		{
			Name:             "Empty list",
			List:             []string{},
			Elements:         []string{"abc", "5"},
			ExpectedElements: []string{"abc", "5"},
		},
		{
			Name:             "Empty elements",
			List:             []string{"1", "def"},
			Elements:         []string{},
			ExpectedElements: []string{},
		},
		{
			Name:             "Elements present in list",
			List:             []string{"1", "def", "3", "abc"},
			Elements:         []string{"3", "abc", "5"},
			ExpectedElements: []string{"5"},
		},
		{
			Name:             "Elements present in unordered list",
			List:             []string{"abc", "def", "1", "3"},
			Elements:         []string{"3", "abc", "5"},
			ExpectedElements: []string{"5"},
		},
		{
			Name:             "Unordered elements present in list",
			List:             []string{"1", "def", "3", "abc"},
			Elements:         []string{"5", "3", "abc"},
			ExpectedElements: []string{"5"},
		},
		{
			Name:             "Unordered elements present in unordered list",
			List:             []string{"abc", "def", "1", "3"},
			Elements:         []string{"5", "3", "abc"},
			ExpectedElements: []string{"5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			res := GetElementsNotInList(tt.List, tt.Elements)

			require.ElementsMatch(t, res, tt.ExpectedElements)
		})
	}
}

func Test_GetElementsInList_uint64(t *testing.T) {

	tests := []struct {
		Name             string
		List             []uint64
		Elements         []uint64
		ExpectedElements []uint64
	}{
		{
			Name:             "Empty list and elements",
			List:             []uint64{},
			Elements:         []uint64{},
			ExpectedElements: []uint64{},
		},
		{
			Name:             "Empty list",
			List:             []uint64{},
			Elements:         []uint64{4, 5},
			ExpectedElements: []uint64{},
		},
		{
			Name:             "Empty elements",
			List:             []uint64{1, 2},
			Elements:         []uint64{},
			ExpectedElements: []uint64{},
		},
		{
			Name:             "Elements present in list",
			List:             []uint64{1, 2, 3, 4},
			Elements:         []uint64{3, 4, 5},
			ExpectedElements: []uint64{3, 4},
		},
		{
			Name:             "Elements present in unordered list",
			List:             []uint64{4, 2, 1, 3},
			Elements:         []uint64{3, 4, 5},
			ExpectedElements: []uint64{3, 4},
		},
		{
			Name:             "Unordered elements present in list",
			List:             []uint64{1, 2, 3, 4},
			Elements:         []uint64{5, 3, 4},
			ExpectedElements: []uint64{3, 4},
		},
		{
			Name:             "Unordered elements present in unordered list",
			List:             []uint64{4, 2, 1, 3},
			Elements:         []uint64{5, 3, 4},
			ExpectedElements: []uint64{4, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			res := GetElementsInList(tt.List, tt.Elements)

			require.ElementsMatch(t, res, tt.ExpectedElements)
		})
	}
}

func Test_GetElementsInList_string(t *testing.T) {

	tests := []struct {
		Name             string
		List             []string
		Elements         []string
		ExpectedElements []string
	}{
		{
			Name:             "Empty list and elements",
			List:             []string{},
			Elements:         []string{},
			ExpectedElements: []string{},
		},
		{
			Name:             "Empty list",
			List:             []string{},
			Elements:         []string{"abc", "5"},
			ExpectedElements: []string{},
		},
		{
			Name:             "Empty elements",
			List:             []string{"1", "def"},
			Elements:         []string{},
			ExpectedElements: []string{},
		},
		{
			Name:             "Elements present in list",
			List:             []string{"1", "def", "3", "abc"},
			Elements:         []string{"3", "abc", "5"},
			ExpectedElements: []string{"3", "abc"},
		},
		{
			Name:             "Elements present in unordered list",
			List:             []string{"abc", "def", "1", "3"},
			Elements:         []string{"3", "abc", "5"},
			ExpectedElements: []string{"3", "abc"},
		},
		{
			Name:             "Unordered elements present in list",
			List:             []string{"1", "def", "3", "abc"},
			Elements:         []string{"5", "3", "abc"},
			ExpectedElements: []string{"3", "abc"},
		},
		{
			Name:             "Unordered elements present in unordered list",
			List:             []string{"abc", "def", "1", "3"},
			Elements:         []string{"5", "3", "abc"},
			ExpectedElements: []string{"abc", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			res := GetElementsInList(tt.List, tt.Elements)

			require.ElementsMatch(t, res, tt.ExpectedElements)
		})
	}
}

func Test_ArrayToLookupMap_uint64(t *testing.T) {
	for _, tt := range []struct {
		IDs            []uint64
		ExpectedResult map[uint64]struct{}
	}{
		{
			IDs:            []uint64{1, 2, 3},
			ExpectedResult: map[uint64]struct{}{1: {}, 2: {}, 3: {}}},
		{
			IDs:            []uint64{},
			ExpectedResult: map[uint64]struct{}{}},
	} {

		result := ArrayToLookupMap(tt.IDs)

		require.Equal(t, tt.ExpectedResult, result)
	}
}

func Test_ArrayToLookupMap_string(t *testing.T) {
	for _, tt := range []struct {
		testStrings    []string
		expectedResult map[string]struct{}
	}{
		{
			testStrings:    []string{"1", "2", "3"},
			expectedResult: map[string]struct{}{"1": {}, "2": {}, "3": {}},
		},
		{
			testStrings:    []string{},
			expectedResult: map[string]struct{}{},
		},
	} {

		result := ArrayToLookupMap(tt.testStrings)

		require.Equal(t, tt.expectedResult, result)
	}
}

func Test_HandleDeprecatedListParam(t *testing.T) {
	for _, tt := range []struct {
		current    []uint64
		deprecated []uint64
		expected   []uint64
	}{
		{ // Appends
			current:    []uint64{1, 2, 3},
			deprecated: []uint64{4, 5, 6},
			expected:   []uint64{1, 2, 3, 4, 5, 6},
		},
		{ // Can use current only
			current:    []uint64{1},
			deprecated: []uint64{},
			expected:   []uint64{1},
		},
		{ // Can use deprecated only
			current:    []uint64{},
			deprecated: []uint64{12},
			expected:   []uint64{12},
		},
		{ // Removes duplicates
			current:    []uint64{1, 1, 2},
			deprecated: []uint64{2, 3},
			expected:   []uint64{1, 2, 3},
		},
		{ // Does not reorder result
			current:    []uint64{1, 5, 10},
			deprecated: []uint64{6, 2, 11},
			expected:   []uint64{1, 5, 10, 6, 2, 11},
		},
	} {
		require.Equal(t, tt.expected, HandleDeprecatedListParam(tt.current, tt.deprecated))
	}
}

func Test_SprintfList_string(t *testing.T) {
	for _, tt := range []struct {
		testStrings    []string
		expectedResult []string
	}{
		{
			testStrings:    []string{"1", "2", "3"},
			expectedResult: []string{"1", "2", "3"},
		},
		{
			testStrings:    []string{"1"},
			expectedResult: []string{"1"},
		},
		{
			testStrings:    []string{},
			expectedResult: []string{},
		},
	} {
		result := SprintfList(tt.testStrings)
		require.Equal(t, tt.expectedResult, result)
	}
}

func Test_SprintfList_int(t *testing.T) {
	for _, tt := range []struct {
		testStrings    []int
		expectedResult []string
	}{
		{
			testStrings:    []int{1, 2, 3},
			expectedResult: []string{"1", "2", "3"},
		},
	} {
		result := SprintfList(tt.testStrings)
		require.Equal(t, tt.expectedResult, result)
	}
}
