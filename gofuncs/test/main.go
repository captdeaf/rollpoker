// This runs all the tests I have for rollpoker
package main

import (
	"sort"
	"fmt"

	// "deafcode.com/rollpoker"
)

type Test struct {
	Name	string
	Fun	func() bool
}

var ALL_TESTS []Test

func RegisterTest(name string, fun func() bool) {
	ALL_TESTS = append(ALL_TESTS, Test{name, fun})
}

func init() {
	RegisterTest("dummy true is true", func() bool {
		return true
	})
}

func main() {
	sort.Slice(ALL_TESTS, func(i, j int) bool { return ALL_TESTS[i].Name < ALL_TESTS[j].Name })

	for _, test := range ALL_TESTS {
		fmt.Printf("Test %s: %v\n", test.Name, test.Fun())
	}
}
