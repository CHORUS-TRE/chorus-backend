package common

import "fmt"

func IsSublist[T comparable](list, sublist []T) bool {
	set := make(map[T]bool)
	for _, value := range list {
		set[value] = true
	}

	for _, value := range sublist {
		if _, found := set[value]; !found {
			return false
		}
	}

	return true
}

func RemoveDuplicates[T comparable](list []T) []T {
	set := make(map[T]bool)
	res := []T{}

	for _, value := range list {
		if _, exists := set[value]; exists {
			continue
		}
		set[value] = true
		res = append(res, value)
	}

	return res
}

// RemoveCommonElementsInLists removes the common elements that are in both lists
// ex. RemoveCommonElementsInLists([A,B,C], [B,D,E]) will return [A,C], [D,E]
func RemoveCommonElementsInLists[T comparable](listA, listB []T) ([]T, []T) {
	set := make(map[T]int)
	for _, value := range RemoveDuplicates(listA) {
		set[value] += 1
	}
	for _, value := range RemoveDuplicates(listB) {
		set[value] += 1
	}

	resA := []T{}
	for _, value := range listA {
		if counter := set[value]; counter > 1 {
			continue
		}
		resA = append(resA, value)
	}

	resB := []T{}
	for _, value := range listB {
		if counter := set[value]; counter > 1 {
			continue
		}
		resB = append(resB, value)
	}

	return resA, resB
}

// AreUnorderedListsEqual compares two lists and check they have the same elements. The order has no importance
// ex. AreUnorderedListsEqual([A,B,C], [B,D,E]) will return false
// ex. AreUnorderedListsEqual([A,B,C], [A,B,C]) will return true
func AreUnorderedListsEqual[T comparable](listA, listB []T) bool {
	cleanedA, cleanedB := RemoveCommonElementsInLists(listA, listB)
	return len(cleanedA) == 0 && len(cleanedB) == 0
}

// GetElementsNotInList returns the elements that are NOT already in the list
// ex. GetElementsNotInList([A,B,C], [B,D,E]) will return [D,E]
func GetElementsNotInList[T comparable](list, elements []T) []T {
	set := make(map[T]bool)
	for _, value := range list {
		set[value] = true
	}

	res := []T{}
	for _, value := range elements {
		if _, found := set[value]; !found {
			res = append(res, value)
		}
	}

	return res
}

// GetElementsInList returns the elements that are already in the list
// ex. GetElementsInList([A,B,C], [B,D,E]) will return [B]
func GetElementsInList[T comparable](list, elements []T) []T {
	set := make(map[T]bool)
	for _, value := range list {
		set[value] = true
	}

	res := []T{}
	for _, value := range elements {
		if _, found := set[value]; found {
			res = append(res, value)
		}
	}

	return res
}

func Batch[T any](batchSize int, elements []T) [][]T {
	var batches [][]T
	for {
		n := len(elements)
		if n > batchSize {
			elements, batches = elements[batchSize:n], append(batches, elements[:batchSize])
		} else {
			batches = append(batches, elements)
			break
		}
	}
	return batches
}

// ArrayToLookupMap provides more efficient 'contains' lookups than repeatedly searching the array
func ArrayToLookupMap[T comparable](values []T) map[T]struct{} {
	res := map[T]struct{}{}
	for _, key := range values {
		res[key] = struct{}{}
	}

	return res
}

func Filter[T any](ts []T, fn func(T) bool) []T {
	result := []T{}
	for _, t := range ts {
		if ok := fn(t); ok {
			result = append(result, t)
		}
	}
	return result
}

func Apply[T any, U any](ts []T, fn func(T) U) []U {
	result := []U{}
	for _, t := range ts {
		result = append(result, fn(t))
	}
	return result
}

func Reverse[T any](ts []T) []T {
	result := []T{}
	for i := range ts {
		result = append(result, ts[len(ts)-1-i])
	}
	return result
}

// HandleDeprecatedListParam provides consistent handling of a list param that has been deprecated and is being replaced
func HandleDeprecatedListParam[T comparable](current []T, deprecated []T) []T {
	return RemoveDuplicates(append(current, deprecated...))
}

func SprintfList[T any](ts []T) []string {
	ret := make([]string, len(ts))
	for i, t := range ts {
		ret[i] = fmt.Sprintf("%v", t)
	}
	return ret
}
