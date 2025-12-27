package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	address, err := net.ResolveUDPAddr("udp", "localhost:42069")

	if err != nil {
		fmt.Println("Error starting UDP connection:", err)
		return
	}

	conn, err := net.DialUDP(address.Network(), nil, address)

	if err != nil {
		fmt.Println("Error dialing UDP connection:", err)
		return
	}

	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')

		if err != nil {
			fmt.Println(err)
			continue
		}

		_, err = conn.Write([]byte(input))

		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}
