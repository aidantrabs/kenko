package redisstore_test

import (
	"fmt"

	"github.com/aidantrabs/kenko/redisstore"
)

func ExampleNew() {
	store := redisstore.New("localhost:6379",
		redisstore.WithPassword("secret"),
		redisstore.WithKeyPrefix("myapp:health"),
	)

	fmt.Println(store != nil)
	// Output: true
}
