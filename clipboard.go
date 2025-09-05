package femto

import (
	"github.com/atotto/clipboard"
)

var internalClipboard string

func clipboardReadAll() (string, error) {
	if clipboard.Unsupported {
		return internalClipboard, nil
	}
	return clipboard.ReadAll()
}

func clipboardWriteAll(s string) error {
	if clipboard.Unsupported {
		internalClipboard = s
		return nil
	}
	return clipboard.WriteAll(s)
}
