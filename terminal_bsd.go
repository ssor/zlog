// +build darwin freebsd openbsd netbsd dragonfly
// +build !appengine

package zlog

import "syscall"

const ioctlReadTermios = syscall.TIOCGETA

type Termios syscall.Termios
