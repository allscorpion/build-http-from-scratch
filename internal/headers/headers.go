package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func (h Headers) Set(key string, value string) {
	currentVal := h.Get(key)

	if currentVal == "" {
		currentVal = value
	} else {
		currentVal = fmt.Sprintf("%v, %v", currentVal, value)
	}

	h[strings.ToLower(key)] = currentVal
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	newLineIndex := bytes.Index(data, []byte("\r\n"))

	if newLineIndex == -1 {
		return 0, false, nil
	}

	if newLineIndex == 0 {
		return 2, true, nil
	}

	currentLine := strings.TrimSpace(string(data[:newLineIndex]))
	keyColon := strings.Index(currentLine, ":")

	if keyColon == -1 {
		return 0, false, fmt.Errorf("malformed request-line: %v", currentLine)
	}

	key := currentLine[:keyColon]

	if strings.Contains(key, " ") {
		return 0, false, fmt.Errorf("spaces in request-line key: %v", key)
	}

	allowedSpecials := "!#$%&'*+-.^_`|~"
	for _, r := range key {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			continue
		}
		if strings.ContainsRune(allowedSpecials, r) {
			continue
		}
		return 0, false, fmt.Errorf("invalid character %q in header key: %s", r, key)
	}

	value := strings.TrimSpace(currentLine[keyColon+1:])
	h.Set(key, value)

	return len(data[:newLineIndex]) + 2, false, nil
}

func NewHeaders() Headers {
	return Headers{}
}
