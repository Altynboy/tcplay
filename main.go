package main

import (
	"log"
	"tcplay/core"
)

func main() {
	conn, err := core.CreateConnection(42069, [4]byte{127, 0, 0, 1})
	if err != nil {
		log.Fatalf("failed to create connection: %v", err)
	}

	log.Println("Raw socked created!")

	err = conn.Connect()
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	log.Println("Connected!")

	err = conn.SendMessage([]byte("Salam broo!"))
	if err != nil {
		log.Fatalf("failed to send msg: %v", err)
	}

	log.Println("Msg send!")
	defer conn.Close()
}
