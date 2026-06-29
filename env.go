package keyring

import (
	"os"
	"strings"
)

// familyPrefix is the env namespace shared by every agent-* CLI and lin: a key K
// is reachable family-wide as familyPrefix + "_" + K (e.g. LIB_AGENT_NO_KEYCHAIN).
// This is the family convention, kept identical to lib-agent-cli's env package so
// the opt-out variable names don't change when a CLI moves to this lib.
const familyPrefix = "LIB_AGENT"

// envNamespace resolves env vars for one CLI: a key is looked up under the CLI's
// own prefix first, then the family-wide fallback. The zero value (empty Prefix)
// consults only the family fallback.
type envNamespace struct {
	// Prefix is the SCREAMING_SNAKE token for this CLI, e.g. "AGENT_SLACK".
	Prefix string
}

// lookup returns the first set of {Prefix}_{key} or familyPrefix_{key}, and
// whether either was present. The specific var wins on presence, so an
// empty-but-set specific var still shadows the family var.
func (n envNamespace) lookup(key string) (string, bool) {
	if n.Prefix != "" {
		if v, ok := os.LookupEnv(n.Prefix + "_" + key); ok {
			return v, true
		}
	}
	return os.LookupEnv(familyPrefix + "_" + key)
}

// flag reports key as a boolean: true for any present value other than "", "0",
// or "false" (case-insensitive). Absent → false.
func (n envNamespace) flag(key string) bool {
	v, ok := n.lookup(key)
	if !ok {
		return false
	}
	switch strings.ToLower(v) {
	case "", "0", "false":
		return false
	default:
		return true
	}
}

// prefixFromName turns a CLI/binary name into an env prefix: uppercase, with
// "-", ".", and " " replaced by "_". So "agent-slack" → "AGENT_SLACK".
func prefixFromName(name string) string {
	r := strings.NewReplacer("-", "_", ".", "_", " ", "_")
	return strings.ToUpper(r.Replace(name))
}

// lastSegment returns the substring after the final "." in s (s itself if there
// is none) — the binary name in a "app.paulie.<name>" service id.
func lastSegment(s string) string {
	if i := strings.LastIndex(s, "."); i >= 0 {
		return s[i+1:]
	}
	return s
}
