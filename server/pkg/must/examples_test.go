package must_test

import (
	"fmt"
	"strconv"

	"github.com/DaanV2/mechanus/server/pkg/must"
)

func ExampleDo() {
	// Convert a string to int, panicking on error
	result := must.Do(strconv.Atoi("42"))
	fmt.Println(result)
	// Output: 42
}

func ExampleDo_multipleValues() {
	// must.Do can be used with any function returning (T, error)
	parseAndDouble := func(s string) (int, error) {
		val, err := strconv.Atoi(s)
		if err != nil {
			return 0, err
		}

		return val * 2, nil
	}

	result := must.Do(parseAndDouble("21"))
	fmt.Println(result)
	// Output: 42
}
