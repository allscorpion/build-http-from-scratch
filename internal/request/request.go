package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type RequestState int

const (
	initialized RequestState = iota
	done        RequestState = iota
)

type Request struct {
	RequestLine RequestLine
	State       RequestState
}

func (r *Request) parse(data []byte) (int, error) {
	if r.State == initialized {
		requestLine, read, err := parseRequestLine(data)

		if err != nil {
			return 0, err
		}

		if read == 0 {
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.State = done

		return read, nil
	}

	if r.State == done {
		return 0, errors.New("trying to read data in a done state")
	}

	return 0, errors.New("unknown state")
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	readToIndex := 0
	buffer := make([]byte, bufferSize)
	request := &Request{
		RequestLine: RequestLine{},
		State:       initialized,
	}
	for request.State != done {
		if readToIndex >= len(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		n, err := reader.Read(buffer[readToIndex:])

		if err != nil {
			if err == io.EOF {
				request.State = done
				break
			}
			return nil, err
		}

		readToIndex += n

		numBytesParsed, err := request.parse(buffer[:readToIndex])

		if err != nil {
			return nil, err
		}

		copy(buffer, buffer[numBytesParsed:])
		readToIndex -= numBytesParsed

	}

	return request, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, idx + 2, nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", str)
	}

	method := parts[0]

	if !isValidRequestLineMethod(method) {
		return nil, fmt.Errorf("malformed method")
	}

	requestTarget := parts[1]

	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", str)
	}

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := versionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", version)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   versionParts[1],
	}, nil
}

func isValidRequestLineMethod(method string) bool {
	for _, char := range method {
		if !unicode.IsUpper(char) && unicode.IsLetter(char) {
			return false
		}
	}

	return true
}
