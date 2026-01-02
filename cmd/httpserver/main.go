package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/allscorpion/build-http-from-scratch/internal/request"
	"github.com/allscorpion/build-http-from-scratch/internal/response"
	"github.com/allscorpion/build-http-from-scratch/internal/server"
)

const port = 42069

func generateHtml(statusCode response.StatusCode, statusText string, header string, body string) string {
	return fmt.Sprintf(`<html>
		<head>
			<title>%v %v</title>
		</head>
		<body>
			<h1>%v</h1>
			<p>%v</p>
		</body>
		</html>
	`, statusCode, statusText, header, body)
}

func writeInternalServerError(w *response.Writer) {
	w.WriteStatusLine(response.InternalServerErrorStatus)
	body := generateHtml(response.InternalServerErrorStatus, "Internal Server Error", "Internal Server Error", "Okay, you know what? This one is on me.")
	headers := response.GetDefaultHeaders(len(body))
	headers.Overwrite("content-type", "text/html")
	w.WriteHeaders(headers)
	w.WriteBody(body)
}

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		if after, ok := strings.CutPrefix(req.RequestLine.RequestTarget, "/httpbin/"); ok {
			urlPath := after
			fullUrl := fmt.Sprintf("https://httpbin.org/%v", urlPath)

			resp, err := http.Get(fullUrl)

			if err != nil {
				writeInternalServerError(w)
				return
			}

			defer resp.Body.Close()

			w.WriteStatusLine(response.OKStatus)
			headers := response.GetDefaultHeaders(0)
			headers.Delete("content-length")
			headers.Set("transfer-encoding", "chunked")
			w.WriteHeaders(headers)

			buffer := make([]byte, 1024)
			for {
				bytesRead, err := resp.Body.Read(buffer)

				fmt.Printf("%v bytes read\n", bytesRead)

				if err != nil {
					if err == io.EOF {
						fmt.Println("reached the end of the file")
						break
					}
					fmt.Printf("an error has occured reading: %v\n", err)
					continue
				}

				data := buffer[:bytesRead]

				bytesWritten, err := w.WriteChunkedBody(data)

				if err != nil {
					fmt.Printf("an error has occured writing: %v\n", err)
					continue
				}

				fmt.Printf("%v bytes written\n", bytesWritten)
			}

			w.WriteChunkedBodyDone()
			return
		}

		if req.RequestLine.RequestTarget == "/yourproblem" {
			w.WriteStatusLine(response.BadRequestStatus)
			body := generateHtml(response.BadRequestStatus, "Bad Request", "Bad Request", "Your request honestly kinda sucked.")
			headers := response.GetDefaultHeaders(len(body))
			headers.Overwrite("content-type", "text/html")
			w.WriteHeaders(headers)
			w.WriteBody(body)
			return
		}

		if req.RequestLine.RequestTarget == "/myproblem" {
			writeInternalServerError(w)
			return
		}

		w.WriteStatusLine(response.OKStatus)
		body := generateHtml(response.OKStatus, "OK", "Success!", "Your request was an absolute banger.")
		headers := response.GetDefaultHeaders(len(body))
		headers.Overwrite("content-type", "text/html")
		w.WriteHeaders(headers)
		w.WriteBody(body)
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
