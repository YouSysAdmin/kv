// Provides dotenv formatting helpers.
//
// Supports two output modes:
//   - DotenvEscaped:    newline characters become literal "\n" (single-line format)
//   - DotenvMultiline:  preserves real newline characters inside quoted values (Ruby Dotenv format for gem version >=3)
//
// IMPORTANT:
//   - Keys are normalized to UPPERCASE; any non-alphanumeric characters are replaced with '_'.
//   - Keys are sorted lexicographically for stable output.
//   - Empty values are printed as KEY= (valid dotenv syntax).
package utils

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/yousysadmin/kv/internal/models"
)

// DotenvMode defines the output style for dotenv formatting.
type DotenvMode string

const (
	// DotenvEscaped converts real newlines to literal "\n"
	// and escapes quotes/backslashes. Most dotenv parsers expect this form.
	DotenvEscaped DotenvMode = "escaped"

	// DotenvMultiline preserves real newline characters inside quotes,
	// compatible with frameworks like Rails dotenv gem.
	DotenvMultiline DotenvMode = "multiline"
)

// ToDotenvMode builds dotenv output using the given mode.
// If withValues is false, prints only keys.
func ToDotenvMode(entities []models.Entity, withValues bool, mode DotenvMode) (string, error) {
	switch mode {
	case DotenvEscaped:
		return toDotenvEscaped(entities, withValues), nil
	case DotenvMultiline:
		return toDotenvMultiline(entities, withValues), nil
	default:
		// Default to escaped mode
		return toDotenvEscaped(entities, withValues), nil
	}
}

// ToDotenv is a shorthand wrapper that defaults to DotenvEscaped mode.
func ToDotenv(entities []models.Entity, withValues bool) (string, error) {
	return ToDotenvMode(entities, withValues, DotenvEscaped)
}

// toDotenvEscaped creates classic dotenv output where every value
// is a single line and real newlines are replaced by literal "\n".
func toDotenvEscaped(entities []models.Entity, withValues bool) string {
	var b strings.Builder
	keys, emap := stableMap(entities, withValues)

	for _, k := range keys {
		v := emap[k]
		v = strings.ReplaceAll(v, "\r\n", "\n")
		v = strings.ReplaceAll(v, `\`, `\\`)
		v = strings.ReplaceAll(v, `"`, `\"`)
		v = strings.ReplaceAll(v, "\n", `\n`)

		if withValues {
			if v == "" {
				fmt.Fprintf(&b, "%s=\n", k)
			} else {
				fmt.Fprintf(&b, "%s=\"%s\"\n", k, v)
			}
		} else {
			fmt.Fprintf(&b, "%s=\n", k)
		}
	}
	return b.String()
}

// toDotenvMultiline preserves real newlines and wraps non-empty values in quotes.
func toDotenvMultiline(entities []models.Entity, withValues bool) string {
	var b strings.Builder
	keys, emap := stableMap(entities, withValues)

	for _, k := range keys {
		v := emap[k]
		v = strings.ReplaceAll(v, "\r\n", "\n")

		if withValues {
			if v == "" {
				fmt.Fprintf(&b, "%s=\n", k)
				continue
			}
			var sb strings.Builder
			sb.Grow(len(v) + 2)
			sb.WriteByte('"')
			for _, r := range v {
				switch r {
				case '\\':
					sb.WriteString(`\\`)
				case '"':
					sb.WriteString(`\"`)
				default:
					sb.WriteRune(r)
				}
			}
			sb.WriteByte('"')
			fmt.Fprintf(&b, "%s=%s\n", k, sb.String())
		} else {
			fmt.Fprintf(&b, "%s=\n", k)
		}
	}
	return b.String()
}

var nonAlnum = regexp.MustCompile(`[^A-Za-z0-9]+`)

// normalizeKey converts a key to uppercase, replaces any non-alphanumeric
// characters with underscores, and trims leading/trailing underscores.
func normalizeKey(k string) string {
	k = strings.ToUpper(k)
	k = nonAlnum.ReplaceAllString(k, "_")
	return strings.Trim(k, "_")
}

// stableMap normalizes keys, collects values depending on withValues flag,
// and returns a sorted slice of keys along with the final map.
func stableMap(entities []models.Entity, withValues bool) ([]string, map[string]string) {
	emap := make(map[string]string, len(entities))
	for _, e := range entities {
		k := normalizeKey(e.Key)
		if k == "" {
			k = "_"
		}
		if withValues {
			emap[k] = e.Value
		} else {
			emap[k] = ""
		}
	}
	keys := make([]string, 0, len(emap))
	for k := range emap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, emap
}
