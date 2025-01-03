package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("error listening tcp port 42069: %s", err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print("Error accepting msg:", err)
			continue
		}

		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		log.Print("Error reading data:", err)
		return
	}

	fmt.Printf("Received: %s\n", buf)
}
