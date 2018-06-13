// +build appengine

package zlog

// IsTerminal returns true if stderr's file descriptor is a terminal.
func IsTerminal() bool {
	return true
}
