package namedstats

import (
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	const testdata = "testdata/named_stats.txt"

	stats, err := Load(testdata)
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	content, err := os.ReadFile(testdata)
	if err != nil {
		t.Fatalf("reading testdata: %v", err)
	}

	if stats.String()+"\n" != string(content) {
		t.Fatalf("comparison between raw testdata and parsed testdata failed")
	}
}

func TestLookup(t *testing.T) {
	var expected uint64 = 57252997
	stats, err := Load("testdata/named_stats.txt")
	if err != nil {
		t.Fatalf("parsing failed: %v", err)
	}

	value, err := stats.Lookup("Name Server Statistics", "UDP queries received")
	if err != nil {
		t.Fatalf("lookup failed: %v", err)
	}

	if value != expected {
		t.Fatalf("bad value: gott %d expected %d", value, expected)
	}
}
