package sectiontrace

import "fmt"

var (
	namesSeen = map[string]bool{}
)

func maybeCheckName(name string) error {
	if !DebugMode {
		return nil
	}
	_, present := namesSeen[name]
	if present {
		return fmt.Errorf("Section name %q reused", name)
	}
	return nil
}
