package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/allscorpion/build-http-from-scratch/internal/headers"
)

type RequestState int

const (
	requestStateInitialized RequestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       RequestState
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return n, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = requestStateParsingBody
		}
		return n, nil
	case requestStateParsingBody:
		contentLength, exists := r.Headers.Get("content-length")

		if !exists {
			r.state = requestStateDone
			return 0, nil
		}

		contentLengthNum, err := strconv.Atoi(contentLength)

		if err != nil {
			return 0, err
		}

		r.Body = append(r.Body, data...)

		bodyLength := len(r.Body)

		if bodyLength > contentLengthNum {
			return 0, fmt.Errorf("the body is larger than the content-length. Received %v, expected %v", bodyLength, contentLengthNum)
		}

		if bodyLength == contentLengthNum {
			r.state = requestStateDone
		}

		return len(data), nil

	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("unknown state")
	}
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
		Headers:     headers.NewHeaders(),
		Body:        make([]byte, 0),
		state:       requestStateInitialized,
	}
	for request.state != requestStateDone {
		if readToIndex >= len(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		numBytesRead, err := reader.Read(buffer[readToIndex:])

		if err != nil {
			if err == io.EOF {
				if request.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", request.state, numBytesRead)
				}
				break
			}
			return nil, err
		}

		readToIndex += numBytesRead

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
	newLineIndex := bytes.Index(data, []byte("\r\n"))
	if newLineIndex == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:newLineIndex])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, newLineIndex + 2, nil
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
