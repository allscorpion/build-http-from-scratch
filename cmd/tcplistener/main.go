package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

func getLinesChannel(file io.ReadCloser) <-chan string {
	ch := make(chan string)
	currentLine := ""

	go func() {
		defer close(ch)
		defer file.Close()
		for {
			buffer := make([]byte, 8)
			n, err := file.Read(buffer)

			if err != nil {
				if err != io.EOF {
					fmt.Println("error reading file:", err)
				}

				break
			}

			data := string(buffer[:n])

			parts := strings.Split(data, "\n")

			for i := 0; i < len(parts)-1; i++ {
				currentLine += parts[i]
				ch <- currentLine
				currentLine = ""
			}

			currentLine += parts[len(parts)-1]
		}

		if currentLine != "" {
			ch <- currentLine
		}
	}()

	return ch
}

func main() {
	network, err := net.Listen("tcp", ":42069")

	if err != nil {
		fmt.Println("Error starting TCP connection:", err)
		return
	}

	defer func() {
		network.Close()
		fmt.Println("The connection has been closed")
	}()

	for {
		conn, err := network.Accept()

		if err != nil {
			fmt.Println("Error accepting TCP connection:", err)
			return
		}

		fmt.Println("The connection has been accepted")
		channel := getLinesChannel(conn)

		for v := range channel {
			fmt.Printf("%v\n", v)
		}
	}

}
