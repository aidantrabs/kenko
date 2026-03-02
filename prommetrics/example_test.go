package prommetrics_test

import (
	"fmt"

	"github.com/aidantrabs/kenko/prommetrics"
	"github.com/prometheus/client_golang/prometheus"
)

func ExampleNew() {
	reporter := prommetrics.New(
		prommetrics.WithRegistry(prometheus.NewRegistry()),
		prommetrics.WithNamespace("myapp"),
	)

	fmt.Println(reporter != nil)
	// Output: true
}
