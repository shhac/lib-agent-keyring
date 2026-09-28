package keyring

import (
	"errors"
	"testing"
)

// A locked store is never touched: reads report not found and writes refuse,
// so no call can put an unlock prompt in front of the user.
func TestLockedStoreIsNotTouched(t *testing.T) {
	k, fb := withFake("app.test.locked", true)
	fb.store["acct"] = "secret"
	fb.locked = true
	if k.Status() != Locked {
		t.Fatalf("status %v", k.Status())
	}
	if v, ok := k.Get("acct"); ok || v != "" {
		t.Fatalf("a locked store was read: %q", v)
	}
	for name, err := range map[string]error{"set": k.Set("acct", "x"), "delete": k.Delete("acct"), "delete all": k.DeleteAll()} {
		if !errors.Is(err, ErrLocked) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if fb.store["acct"] != "secret" {
		t.Fatal("a locked store was written")
	}
	fb.locked = false
	if v, ok := k.Get("acct"); !ok || v != "secret" || k.Status() != Ready {
		t.Fatalf("an unlocked store: %q %v %v", v, ok, k.Status())
	}
}

func TestOptedOutStoreIsUnavailable(t *testing.T) {
	k, _ := withFake("app.test.status", true)
	t.Setenv(NoKeychainEnv, "1")
	if k.Status() != Unavailable || !errors.Is(k.Set("a", "b"), ErrUnavailable) {
		t.Fatalf("status %v", k.Status())
	}
}

func TestStatusNames(t *testing.T) {
	for s, want := range map[Status]string{Ready: "ready", Locked: "locked", Unavailable: "unavailable"} {
		if s.String() != want {
			t.Errorf("%d: %s", s, s)
		}
	}
}
