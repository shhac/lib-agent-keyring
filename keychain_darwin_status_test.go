//go:build darwin

package keyring

import (
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func searchList(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("security", "list-keychains", "-d", "user").Output()
	if err != nil {
		t.Skipf("security list-keychains: %v", err)
	}
	return string(out)
}

// The status is read from a throwaway keychain, never the user's: it must see
// the lock come and go, answer at once, and ask nothing. A prompt would block
// the call, so a slow answer fails the test.
func TestKeychainStatusSeesTheLockWithoutAsking(t *testing.T) {
	before := searchList(t)
	path := filepath.Join(t.TempDir(), "throwaway.keychain-db")
	const password = "throwaway-password"
	if out, err := exec.Command("security", "create-keychain", "-p", password, path).CombinedOutput(); err != nil {
		t.Skipf("create-keychain: %v %s", err, out)
	}
	t.Cleanup(func() { _ = exec.Command("security", "delete-keychain", path).Run() })
	if searchList(t) != before {
		// Locking a keychain other processes search would let them prompt.
		t.Fatal("create-keychain changed the search list; not locking it")
	}
	timed := func() Status {
		t.Helper()
		start := time.Now()
		s := keychainStatusAt(path)
		if elapsed := time.Since(start); elapsed > 2*time.Second {
			t.Fatalf("status took %v: something waited on the user", elapsed)
		}
		return s
	}
	if s := timed(); s != Ready {
		t.Fatalf("a new keychain: %v", s)
	}
	if err := exec.Command("security", "lock-keychain", path).Run(); err != nil {
		t.Fatal(err)
	}
	if s := timed(); s != Locked {
		t.Fatalf("a locked keychain: %v", s)
	}
	if err := exec.Command("security", "unlock-keychain", "-p", password, path).Run(); err != nil {
		t.Fatal(err)
	}
	if s := timed(); s != Ready {
		t.Fatalf("an unlocked keychain: %v", s)
	}
	if s := keychainStatusAt(filepath.Join(t.TempDir(), "missing.keychain-db")); s != Unavailable {
		t.Fatalf("a missing keychain: %v", s)
	}
}

func TestHostStatusAnswersAtOnce(t *testing.T) {
	start := time.Now()
	s := HostStatus()
	if time.Since(start) > 2*time.Second {
		t.Fatal("HostStatus waited")
	}
	t.Logf("host keychain: %v", s)
}
