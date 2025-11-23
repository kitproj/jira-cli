package flag

import (
	"fmt"
	"strings"
)

// FieldFlag is a custom flag type that accumulates multiple field=value pairs
type FieldFlag map[string]string

func (f FieldFlag) String() string {
	return ""
}

func (f FieldFlag) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid field format: %s (expected field=value)", value)
	}
	f[parts[0]] = parts[1]
	return nil
}

