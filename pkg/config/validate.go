package config

import (
	"fmt"
	"strings"
)

func validateLoginNames(names []string, field string) error {
	for _, name := range names {
		if strings.ContainsAny(name, ".*?^+$") {
			return fmt.Errorf("%s contains an invalid character ('.', '*', '?', '^', '+', or '$'): %q. Glob and regular expression patterns are not supported", field, name)
		}
	}
	return nil
}
