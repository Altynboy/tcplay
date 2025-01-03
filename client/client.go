package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:42069")
	if err != nil {
		log.Fatal("Error while creating connection:", err)
	}

	_, err = conn.Write([]byte("Salam bro!"))
	if err != nil {
		fmt.Println(err)
	}

	defer conn.Close()
}
