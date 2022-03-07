package genericache_test

import (
	"fmt"
	"github.com/nathanejohnson/genericache"
)

func fill(i int) (string, error) {
	return fmt.Sprintf("%d => %d", i, i+5), nil
}

func Example() {
	c := genericache.NewGeneriCache(fill, false)
	v1, _ := c.Get(1)
	fmt.Println(v1)
	v2, _ := c.Get(2)
	fmt.Println(v2)
	// Output:
	// 1 => 6
	// 2 => 7
}
