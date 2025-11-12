package xslices_test

import (
	"fmt"

	"github.com/DaanV2/mechanus/server/pkg/extensions/xslices"
)

func ExampleFilter() {
	numbers := []int{1, 2, 3, 4, 5, 6}
	// Filter removes numbers greater than 3
	result := xslices.Filter(numbers, func(n int) bool {
		return n > 3
	})
	fmt.Println(result)
	// Output: [1 2 3]
}

func ExampleFilter_strings() {
	words := []string{"apple", "banana", "apricot", "cherry"}
	// Filter removes strings that start with "a"
	result := xslices.Filter(words, func(s string) bool {
		return s != "" && s[0] == 'a'
	})
	fmt.Println(result)
	// Output: [banana cherry]
}

func ExampleMap() {
	numbers := []int{1, 2, 3, 4, 5}
	// Map transforms integers to their squares
	squares := xslices.Map(numbers, func(n int) int {
		return n * n
	})
	fmt.Println(squares)
	// Output: [1 4 9 16 25]
}

func ExampleMap_stringToInt() {
	words := []string{"cat", "dog", "bird"}
	// Map transforms strings to their lengths
	lengths := xslices.Map(words, func(s string) int {
		return len(s)
	})
	fmt.Println(lengths)
	// Output: [3 3 4]
}

type Item struct {
	ID   string
	Name string
}

func (i Item) GetID() string {
	return i.ID
}

func ExampleContainsID() {
	items := []Item{
		{ID: "1", Name: "First"},
		{ID: "2", Name: "Second"},
		{ID: "3", Name: "Third"},
	}

	fmt.Println(xslices.ContainsID(items, "2"))
	fmt.Println(xslices.ContainsID(items, "5"))
	// Output:
	// true
	// false
}

func ExampleCollectIDs() {
	items := []Item{
		{ID: "a1", Name: "First"},
		{ID: "b2", Name: "Second"},
		{ID: "c3", Name: "Third"},
	}

	ids := xslices.CollectIDs(items)
	fmt.Println(ids)
	// Output: [a1 b2 c3]
}

func ExampleAddIfMissing() {
	items := []Item{
		{ID: "1", Name: "First"},
		{ID: "2", Name: "Second"},
	}

	newItems := []Item{
		{ID: "2", Name: "Second Updated"}, // Already exists
		{ID: "3", Name: "Third"},          // New
	}

	result := xslices.AddIfMissing(items, newItems...)
	fmt.Println(len(result))
	for _, item := range result {
		fmt.Printf("%s: %s\n", item.ID, item.Name)
	}
	// Output:
	// 3
	// 1: First
	// 2: Second
	// 3: Third
}

func ExampleRemoveID() {
	items := []Item{
		{ID: "1", Name: "First"},
		{ID: "2", Name: "Second"},
		{ID: "3", Name: "Third"},
	}

	toRemove := []Item{
		{ID: "2", Name: ""},
	}

	result := xslices.RemoveID(items, toRemove...)
	fmt.Println(len(result))
	for _, item := range result {
		fmt.Printf("%s: %s\n", item.ID, item.Name)
	}
	// Output:
	// 2
	// 1: First
	// 3: Third
}
