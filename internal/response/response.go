package response

import (
	"fmt"
	"io"

	"github.com/allscorpion/build-http-from-scratch/internal/headers"
)

type StatusCode int

const (
	OKStatus                  StatusCode = 200
	BadRequestStatus          StatusCode = 400
	InternalServerErrorStatus StatusCode = 500
)

func getStatusLine(statusCode StatusCode) string {
	switch statusCode {
	case OKStatus:
		return "HTTP/1.1 200 OK"
	case BadRequestStatus:
		return "HTTP/1.1 400 Bad Request"
	case InternalServerErrorStatus:
		return "HTTP/1.1 500 Internal Server Error"
	default:
		return fmt.Sprintf("HTTP/1.1 %v ", statusCode)
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	responseHeaders := headers.NewHeaders()
	responseHeaders.Set("content-length", fmt.Sprint(contentLen))
	responseHeaders.Set("connection", "close")
	responseHeaders.Set("content-type", "text/plain")
	return responseHeaders
}

type Writer struct {
	writer io.Writer
}

func NewWriter(writer io.Writer) Writer {
	return Writer{
		writer: writer,
	}
}

func (w *Writer) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	_, err := w.Write([]byte(getStatusLine(statusCode) + "\r\n"))

	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	for key, value := range headers {
		_, err := fmt.Fprintf(w, "%v: %v\r\n", key, value)

		if err != nil {
			return err
		}
	}

	_, err := fmt.Fprintf(w, "\r\n")

	return err
}

func (w *Writer) WriteBody(body string) error {
	_, err := fmt.Fprintf(w, "%s", body)

	return err
}
