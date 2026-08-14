package envx

import (
	"os"
	"strings"
)

func Merge(extra ...string) []string {
	m := make(map[string]string, len(os.Environ())+len(extra))
	for _, kv := range os.Environ() {
		if i := strings.IndexByte(kv, '='); i > 0 {
			m[strings.ToUpper(kv[:i])] = kv[i+1:]
		}
	}
	for _, kv := range extra {
		if i := strings.IndexByte(kv, '='); i > 0 {
			m[strings.ToUpper(kv[:i])] = kv[i+1:]
		}
	}
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}
