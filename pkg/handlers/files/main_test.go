// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
package files

import (
	"crypto/rand"
	"fmt"
	"os"
	"testing"

	"github.com/azure/peerd/pkg/cache"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	err := teardown()
	if code == 0 && err != nil {
		code = 42
	}
	os.Exit(code)
}

func setup() {
	suf := newRandomStringN(10)
	cache.Path += suf
}

// teardown removes the cache directory.
func teardown() error {
	if err := os.RemoveAll(cache.Path); err != nil {
		return fmt.Errorf("failed to remove cache dir: %v --- %v", cache.Path, err)
	}

	return nil
}

// newRandomStringN creates a new random string of length n.
func newRandomStringN(n int) string {
	randBytes := make([]byte, n/2)
	_, _ = rand.Read(randBytes)

	return fmt.Sprintf("%x", randBytes)
}
