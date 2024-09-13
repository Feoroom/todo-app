package assert

import "testing"

func Equal[T comparable](t *testing.T, actual, expecting T) {
	t.Helper()

	if actual != expecting {
		t.Errorf("got: %v, want: %v", actual, expecting)
	}
}
