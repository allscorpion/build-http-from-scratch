package main

import (
	"crypto/sha256"
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
			h := response.GetDefaultHeaders(0)
			h.Delete("content-length")
			h.Set("transfer-encoding", "chunked")
			h.Set("trailer", "X-Content-Sha256, X-Content-Length")
			w.WriteHeaders(h)

			buffer := make([]byte, 1024)
			fullResponseBody := []byte{}

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
				fullResponseBody = append(fullResponseBody, data...)

				bytesWritten, err := w.WriteChunkedBody(data)

				if err != nil {
					fmt.Printf("an error has occured writing: %v\n", err)
					continue
				}

				fmt.Printf("%v bytes written\n", bytesWritten)
			}

			w.WriteChunkedBodyDone()
			fmt.Println("body has finished being written")
			trailers := map[string]string{}
			trailers["X-Content-Sha256"] = fmt.Sprintf("%x", sha256.Sum256(fullResponseBody))
			trailers["X-Content-Length"] = fmt.Sprintf("%v", len(fullResponseBody))
			// trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", sha256.Sum256(fullResponseBody)))
			// trailers.Set("X-Content-Length", fmt.Sprintf("%v", len(fullResponseBody)))

			w.WriteTrailers(trailers)
			fmt.Println("finished writing trailers")
			return
		}

		if req.RequestLine.RequestTarget == "/video" {
			data, err := os.ReadFile("assets/vim.mp4")

			if err != nil {
				fmt.Println(err)
				writeInternalServerError(w)
				return
			}

			w.WriteStatusLine(response.OKStatus)
			h := response.GetDefaultHeaders(len(data))
			h.Overwrite("content-type", "video/mp4")
			w.WriteHeaders(h)
			w.WriteBody(string(data))

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
