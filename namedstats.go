// Package namedstats provides functions and methods for parsing and
// looking up values from BIND named_stats files
package namedstats

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Stats represents the parsed named_stats file
type Stats struct {
	// Time named_stats.txt was written to file
	Time time.Time `json:"time" yaml:"time"`

	// Section map
	Section      map[string]*Section `json:"sections" yaml:"sections"`
	sectionOrder []string            `json:"-" yaml:"-"`
}

func (s *Stats) String() string {
	epoch := s.Time.Unix()
	var entries []string
	entries = append(entries, fmt.Sprintf("+++ Statistics Dump +++ (%d)", epoch))
	for _, key := range s.sectionOrder {
		entries = append(entries, s.Section[key].String())
	}
	entries = append(entries, fmt.Sprintf("--- Statistics Dump --- (%d)", epoch))
	return strings.Join(entries, "\n")
}

// Section represents a section of the Stats
type Section struct {
	// Section name
	Name string `json:"name" yaml:"name"`

	// SubSection map
	SubSection map[string]*SubSection `json:"sub_sections" yaml:"subSections"`

	// Key/value map
	KeyValue        map[string]uint64 `json:"key_value" yaml:"keyValue"`
	subSectionOrder []string          `json:"-" yaml:"-"`
	keyValueOrder   []string          `json:"-" yaml:"-"`
}

func (se *Section) String() string {
	var entries []string
	entries = append(entries, "++ "+se.Name+" ++")
	for _, key := range se.subSectionOrder {
		entries = append(entries, se.SubSection[key].String())
	}
	for _, key := range se.keyValueOrder {
		entries = append(entries, fmt.Sprintf("%20d %s", se.KeyValue[key], key))
	}

	return strings.Join(entries, "\n")
}

// SubSection represents a subsection of a section
type SubSection struct {
	// SubSection name
	Name string `json:"name" yaml:"name"`

	// Key/value map
	KeyValue      map[string]uint64 `json:"key_value" yaml:"keyValue"`
	keyValueOrder []string          `json:"-" yaml:"-"`
}

func (sub *SubSection) String() string {
	var entries []string
	entries = append(entries, "["+sub.Name+"]")
	for _, key := range sub.keyValueOrder {
		entries = append(entries, fmt.Sprintf("%20d %s", sub.KeyValue[key], key))
	}

	return strings.Join(entries, "\n")
}

// Load reads a named_stats.txt file, and returns a *Stats upon
// success.
func Load(name string) (*Stats, error) {
	content, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}

	return New(content)
}

// New parses named_stats data, and returns a *Stats upon success.
func New(content []byte) (*Stats, error) {
	tm, lines, err := unwrap(content)
	if err != nil {
		return nil, err
	}

	s := &Stats{
		Time:    tm,
		Section: make(map[string]*Section),
	}

	var section *Section
	var subsection *SubSection
	for lineno, line := range lines {
		if name := getSection(line); name != "" {
			section = &Section{
				Name:       name,
				SubSection: make(map[string]*SubSection),
				KeyValue:   make(map[string]uint64),
			}
			s.Section[name] = section
			s.sectionOrder = append(s.sectionOrder, name)
			subsection = nil

			continue
		} else if section == nil {
			return nil, fmt.Errorf("expection section in line %d", lineno+2)
		}
		if name := getSubSection(line); name != "" {
			subsection = &SubSection{
				Name:     name,
				KeyValue: make(map[string]uint64),
			}
			section.SubSection[name] = subsection
			section.subSectionOrder = append(section.subSectionOrder, name)
			continue
		}

		key, value, err := getKeyValue(strings.TrimSpace(line))
		if err != nil {
			return nil, fmt.Errorf("line %d: %v", lineno+2, err)
		}
		if subsection != nil {
			subsection.KeyValue[key] = value
			subsection.keyValueOrder = append(subsection.keyValueOrder, key)
		} else {
			section.KeyValue[key] = value
			section.keyValueOrder = append(section.keyValueOrder, key)
		}
	}

	return s, nil
}

// Lookup looks up a value from a key path. Keys are: section name,
// subsection name, key.
func (s *Stats) Lookup(keys ...string) (uint64, error) {
	if len(keys) == 0 {
		return 0, fmt.Errorf("stats: missing keys")
	}

	key, keys := keys[0], keys[1:]

	if sec, ok := s.Section[key]; ok {
		v, err := sec.lookup(keys...)
		if err != nil {
			return 0, fmt.Errorf("section: %q: %v", key, err)
		}
		return v, nil
	}
	return 0, fmt.Errorf("no section with name: %s", key)
}

func (se *Section) lookup(keys ...string) (uint64, error) {
	if len(keys) == 0 {
		return 0, fmt.Errorf("section: missing keys")
	}

	key, keys := keys[0], keys[1:]
	if sub, ok := se.SubSection[key]; ok {
		return sub.lookup(keys...)
	}
	if v, ok := se.KeyValue[key]; ok {
		return v, nil
	}

	return 0, fmt.Errorf("%q: no match", key)
}

func (sub *SubSection) lookup(keys ...string) (uint64, error) {
	if len(keys) == 0 {
		return 0, fmt.Errorf("subsection: missing keys")
	}

	key := keys[0]
	if v, ok := sub.KeyValue[key]; ok {
		return v, nil
	}

	return 0, fmt.Errorf("kv %q: no match", key)
}

func getKeyValue(s string) (string, uint64, error) {
	number, key, ok := strings.Cut(s, " ")
	if !ok {
		return "", 0, fmt.Errorf("bad val/key: %s", s)
	}

	u, err := strconv.ParseUint(number, 10, 64)
	if err != nil {
		return "", 0, err
	}

	return key, u, nil
}

func getSubSection(s string) string {
	if after, ok := strings.CutPrefix(s, "["); ok {
		if before, ok := strings.CutSuffix(after, "]"); ok {
			return before
		}
	}
	return ""
}

func getSection(s string) string {
	if after, ok := strings.CutPrefix(s, "++ "); ok {
		if before, ok := strings.CutSuffix(after, " ++"); ok {
			return before
		}
	}
	return ""
}

func unwrap(content []byte) (time.Time, []string, error) {
	tm := time.Time{}
	lines := strings.Split(string(content), "\n")
	if len(lines) < 2 {
		return tm, nil, fmt.Errorf("bad format")
	}

	t0, err := readEncap(lines[0], true)
	if err != nil {
		return tm, nil, err
	}

	t1, err := readEncap(lines[len(lines)-2], false)
	if err != nil {
		return tm, nil, err
	}

	if t0 != t1 {
		return tm, nil, fmt.Errorf("container encapsulation time mismatch: %d vs %d", t0, t1)
	}

	tm = time.Unix(t0, 0)

	return tm, lines[1 : len(lines)-2], nil
}

func readEncap(s string, head bool) (int64, error) {
	prefix := "+++ Statistics Dump +++ ("
	if !head {
		prefix = "--- Statistics Dump --- ("
	}

	if after, ok := strings.CutPrefix(s, prefix); ok {
		if before, ok := strings.CutSuffix(after, ")"); ok {
			u, err := strconv.ParseInt(before, 10, 64)
			if err != nil {
				return 0, fmt.Errorf("bad encapsulation: %v", err)
			}
			return u, nil
		}
	}

	return 0, fmt.Errorf("bad encapsulation: %s", s)
}
