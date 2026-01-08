package ui

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
)

// Clipboard provides system clipboard operations.
type Clipboard struct {
	fallbackBuffer string // Used when system clipboard is unavailable
}

// NewClipboard creates a new clipboard instance.
func NewClipboard() *Clipboard {
	return &Clipboard{}
}

// Copy copies text to the system clipboard.
func (c *Clipboard) Copy(text string) error {
	// Store in fallback buffer regardless
	c.fallbackBuffer = text

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		// Try xclip first, then xsel
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else if _, err := exec.LookPath("wl-copy"); err == nil {
			// Wayland support
			cmd = exec.Command("wl-copy")
		} else {
			// No clipboard tool available, use fallback
			return nil
		}
	default:
		// Unsupported platform, use fallback
		return nil
	}

	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// Paste retrieves text from the system clipboard.
func (c *Clipboard) Paste() (string, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbpaste")
	case "linux":
		// Try xclip first, then xsel
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard", "-o")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--output")
		} else if _, err := exec.LookPath("wl-paste"); err == nil {
			// Wayland support
			cmd = exec.Command("wl-paste", "-n")
		} else {
			// No clipboard tool available, use fallback
			return c.fallbackBuffer, nil
		}
	default:
		// Unsupported platform, use fallback
		return c.fallbackBuffer, nil
	}

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		// On error, return fallback buffer
		return c.fallbackBuffer, nil
	}

	return out.String(), nil
}

// HasSystemClipboard returns true if a system clipboard is available.
func HasSystemClipboard() bool {
	switch runtime.GOOS {
	case "darwin":
		_, err := exec.LookPath("pbcopy")
		return err == nil
	case "linux":
		if _, err := exec.LookPath("xclip"); err == nil {
			return true
		}
		if _, err := exec.LookPath("xsel"); err == nil {
			return true
		}
		if _, err := exec.LookPath("wl-copy"); err == nil {
			return true
		}
		return false
	default:
		return false
	}
}
