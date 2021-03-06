// +build darwin,cgo dragonfly,cgo freebsd,cgo linux,cgo nacl,cgo netbsd,cgo openbsd,cgo solaris,cgo

package vt100

import (
	"github.com/pkg/term"
	"time"
)

var (
	defaultTimeout = 10 * time.Millisecond
	lastKey        int
)

type TTY struct {
	t       *term.Term
	timeout time.Duration
}

// NewTTY opens /dev/tty in raw and cbreak mode as a term.Term
func NewTTY() (*TTY, error) {
	t, err := term.Open("/dev/tty", term.RawMode, term.CBreakMode, term.ReadTimeout(defaultTimeout))
	if err != nil {
		return nil, err
	}
	return &TTY{t, defaultTimeout}, nil
}

// Term will return the underlying term.Term
func (tty *TTY) Term() *term.Term {
	return tty.t
}

// RawMode will switch the terminal to raw mode
func (tty *TTY) RawMode() {
	term.RawMode(tty.t)
}

// NoBlock leaves "cooked" mode and enters "cbreak" mode
func (tty *TTY) NoBlock() {
	tty.t.SetCbreak()
}

// SetTimeout sets a timeout for reading a key
func (tty *TTY) SetTimeout(d time.Duration) {
	tty.timeout = d
	tty.t.SetReadTimeout(tty.timeout)
}

// Restore will restore the terminal
func (tty *TTY) Restore() {
	tty.t.Restore()
}

// Close will Restore and close the raw terminal
func (tty *TTY) Close() {
	t := tty.Term()
	t.Restore()
	t.Close()
}

// Thanks https://stackoverflow.com/a/32018700/131264
// Returns either an ascii code, or (if input is an arrow) a Javascript key code.
func asciiAndKeyCode(tty *TTY) (ascii, keyCode int, err error) {
	bytes := make([]byte, 3)
	var numRead int
	tty.RawMode()
	tty.NoBlock()
	tty.SetTimeout(tty.timeout)
	numRead, err = tty.t.Read(bytes)
	tty.Restore()
	tty.t.Flush()
	if err != nil {
		return
	}
	if numRead == 3 && bytes[0] == 27 && bytes[1] == 91 {
		// Three-character control sequence, beginning with "ESC-[".

		// Since there are no ASCII codes for arrow keys, we use
		// Javascript key codes.
		if bytes[2] == 65 {
			// Up
			keyCode = 38
		} else if bytes[2] == 66 {
			// Down
			keyCode = 40
		} else if bytes[2] == 67 {
			// Right
			keyCode = 39
		} else if bytes[2] == 68 {
			// Left
			keyCode = 37
		}
	} else if numRead == 1 {
		ascii = int(bytes[0])
	} else {
		// Two characters read??
	}
	return
}

// Returns either an ascii code, or (if input is an arrow) a Javascript key code.
func asciiAndKeyCodeOnce() (ascii, keyCode int, err error) {
	t, err := NewTTY()
	if err != nil {
		return 0, 0, err
	}
	a, kc, err := asciiAndKeyCode(t)
	t.Close()
	return a, kc, err
}

func (tty *TTY) ASCII() int {
	ascii, _, err := asciiAndKeyCode(tty)
	if err != nil {
		return 0
	}
	return ascii
}

func ASCIIOnce() int {
	ascii, _, err := asciiAndKeyCodeOnce()
	if err != nil {
		return 0
	}
	return ascii
}

func (tty *TTY) KeyCode() int {
	_, keyCode, err := asciiAndKeyCode(tty)
	if err != nil {
		return 0
	}
	return keyCode
}

func KeyCodeOnce() int {
	_, keyCode, err := asciiAndKeyCodeOnce()
	if err != nil {
		return 0
	}
	return keyCode
}

// Return the keyCode or ascii, but ignore repeated keys
func (tty *TTY) Key() int {
	ascii, keyCode, err := asciiAndKeyCode(tty)
	if err != nil {
		lastKey = 0
		return 0
	}
	if keyCode != 0 {
		if keyCode == lastKey {
			lastKey = 0
			return 0
		}
		lastKey = keyCode
		return keyCode
	}
	if ascii == lastKey {
		lastKey = 0
		return 0
	}
	lastKey = ascii
	return ascii
}

func KeyOnce() int {
	ascii, keyCode, err := asciiAndKeyCodeOnce()
	if err != nil {
		return 0
	}
	if keyCode != 0 {
		return keyCode
	}
	return ascii
}

// Wait for Esc, Enter or Space to be pressed
func WaitForKey() {
	// Get a new TTY and start reading keypresses in a loop
	r, err := NewTTY()
	defer r.Close()
	if err != nil {
		panic(err)
	}
	//r.SetTimeout(10 * time.Millisecond)
	for {
		switch r.Key() {
		case 27, 13, 32:
			return
		}
	}
}
