package api

import (
	"net/url"
	"strings"
)

// ParseValues parses a string like `abc=123,foo=bar,name=%42%65%64%72%6f%6f%6d` to url.Values.
func ParseValues(s string) url.Values {
	out := url.Values{}
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		var value string
		if len(parts) == 2 {
			value, _ = url.QueryUnescape(parts[1])
		}
		out.Add(parts[0], value)
	}
	return out
}
