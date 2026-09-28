//go:build darwin

package keyring

import (
	"sync"

	"github.com/ebitengine/purego"
)

// The Security framework's own keychain status, read through purego so the
// package stays CGO-free. SecKeychainGetStatus reads state and never asks the
// user anything, unlike the `security` CLI, whose reads prompt to unlock.
var security struct {
	once        sync.Once
	ok          bool
	copyDefault func(*uintptr) int32
	open        func(string, *uintptr) int32
	getStatus   func(uintptr, *uint32) int32
	release     func(uintptr)
}

// kSecUnlockStateStatus: the keychain is unlocked.
const unlockedBit = 1

func loadSecurity() bool {
	security.once.Do(func() {
		lib, err := purego.Dlopen("/System/Library/Frameworks/Security.framework/Security", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			return
		}
		cf, err := purego.Dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			return
		}
		purego.RegisterLibFunc(&security.copyDefault, lib, "SecKeychainCopyDefault")
		purego.RegisterLibFunc(&security.open, lib, "SecKeychainOpen")
		purego.RegisterLibFunc(&security.getStatus, lib, "SecKeychainGetStatus")
		purego.RegisterLibFunc(&security.release, cf, "CFRelease")
		security.ok = true
	})
	return security.ok
}

// defaultKeychainStatus is the status of the user's default keychain,
// normally the login keychain.
func defaultKeychainStatus() Status {
	if !loadSecurity() {
		return Unavailable
	}
	var keychain uintptr
	if security.copyDefault(&keychain) != 0 || keychain == 0 {
		return Unavailable
	}
	defer security.release(keychain)
	return keychainStatus(keychain)
}

// keychainStatusAt is the status of the keychain file at path.
func keychainStatusAt(path string) Status {
	if !loadSecurity() {
		return Unavailable
	}
	var keychain uintptr
	if security.open(path, &keychain) != 0 || keychain == 0 {
		return Unavailable
	}
	defer security.release(keychain)
	return keychainStatus(keychain)
}

func keychainStatus(keychain uintptr) Status {
	var bits uint32
	if security.getStatus(keychain, &bits) != 0 {
		return Unavailable
	}
	if bits&unlockedBit == 0 {
		return Locked
	}
	return Ready
}
