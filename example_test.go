package namedstats_test

import (
	"fmt"
	"github.com/stianwa/namedstats"
	"log"
)

func Example() {
	stats, err := namedstats.Load("somedir/named_stats.txt")
	if err != nil {
		log.Fatal(err)
	}

	value, err := stats.Lookup("Name Server Statistics", "UDP queries received")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%d\n", value)
}
