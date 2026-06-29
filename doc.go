// Package keyring stores small secrets in the host's OS secret store — the macOS
// login Keychain, the Linux Secret Service, or the Windows Credential Manager —
// keyed by a service id and an account name, with a 0600-file fallback left to
// the caller when no store is available.
//
// It is the shared secret-storage primitive for the agent-* family: lib-agent-cli
// (credentials) and lib-agent-mcp (the OAuth signing key + pairing code) both use
// it, so the OS-specific backends live in exactly one place. The package is
// deliberately small and free of family dependencies.
//
//	kr := keyring.New("app.example.agent-foo")
//	if kr.Available() {
//		_ = kr.Set("token", secret)
//		v, ok := kr.Get("token")
//	}
//
// An OS opt-out env var (AGENT_FOO_NO_KEYCHAIN, or the family-wide
// LIB_AGENT_NO_KEYCHAIN) makes Available report false, so non-interactive hosts
// fall back to file storage instead of blocking on a GUI prompt.
package keyring
