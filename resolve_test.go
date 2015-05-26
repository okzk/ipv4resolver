package ipv4resolver

import (
	"testing"
)

func TestResolve(t *testing.T) {
	_, _, err := ResolveHost("github.com")
	if err != nil {
		t.Error(err)
	}
}
