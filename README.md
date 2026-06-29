# lib-agent-keyring

The agent-* family's shared **OS secret store**: store and read small secrets in
the macOS login Keychain, the Linux Secret Service, or the Windows Credential
Manager, keyed by a service id and account — with a clean "unavailable" signal so
callers can fall back to file storage on hosts without a usable store.

It exists so the OS-specific backend lives in **one** place. `lib-agent-cli`
(API credentials) and `lib-agent-mcp` (the local-OAuth signing key + pairing
code) both depend on it; neither re-implements keychain access, and the module
stays free of family dependencies (only `go-keyring`).

## Usage

```go
import "github.com/shhac/lib-agent-keyring"

kr := keyring.New("app.example.agent-foo") // reverse-DNS service id
if kr.Available() {
    _ = kr.Set("token", secret)   // store
    v, ok := kr.Get("token")      // read
    _ = kr.Delete("token")        // remove
}
```

- **`New(service)`** derives the opt-out env namespace from the service's last
  dotted segment (`app.paulie.agent-slack` → `AGENT_SLACK`). **`NewWithEnvPrefix`**
  sets it explicitly.
- **`Available()`** is false when no OS store is reachable, or when the opt-out
  env var is set — `AGENT_FOO_NO_KEYCHAIN` (per-CLI) or `LIB_AGENT_NO_KEYCHAIN`
  (family-wide). When false, mutations return `ErrUnavailable` so the caller can
  use a file store. This is what keeps credential writes testable in CI and stops
  non-interactive hosts from blocking on a GUI prompt.

## Backends

| OS | Backend |
|---|---|
| macOS | the `security` CLI (raw values, so existing items stay readable) |
| Linux | Secret Service (libsecret) over D-Bus, via `go-keyring` |
| Windows | Credential Manager, via `go-keyring` |
| other | no-op: `Available()` is false → file fallback |

Adding an OS is a new build-tagged file, not a conditional sprinkled through the
methods.

## Develop

```sh
go test ./... -count=1
go vet ./...
golangci-lint run ./...
```

## License

PolyForm Perimeter 1.0.0 — see [LICENSE](LICENSE).
