# namedstats
[![Go Reference](https://pkg.go.dev/badge/github.com/stianwa/namedstats.svg)](https://pkg.go.dev/github.com/stianwa/namedstats) [![Go Report Card](https://goreportcard.com/badge/github.com/stianwa/namedstats)](https://goreportcard.com/report/github.com/stianwa/namedstats)
Package namedstats implements a BIND named_stats.txt parser.

Installation
------------

The recommended way to install namedstats

```
go get github.com/stianwa/namedstats
```

Example
-------

```go

package main

import (
	"fmt"
	"github.com/stianwa/namedstats"
	"log"
	"os"
)

func main() {
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
```

State
-----
The namedstats module is currently under development.

License
-------

GPLv3, see [LICENSE.md](LICENSE.md)
