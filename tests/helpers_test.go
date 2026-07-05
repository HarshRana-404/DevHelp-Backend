package tests

import "go.uber.org/zap"

// newNopLogger returns a no-op Zap logger suitable for tests.
// It discards all log output so test output stays clean.
func newNopLogger() *zap.Logger {
	return zap.NewNop()
}
