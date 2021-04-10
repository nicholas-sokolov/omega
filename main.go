package main

import (
	"log"
	"net"

	"github.com/nicholas-sokolov/omega/server"
)

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:8000")
	if err != nil {
		log.Fatal("Can't listen the address")
	}
	defer l.Close()

	log.Println("Start listening...")

	s := server.NewServer()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Can't accept, %s", err)
		}

		if err := s.HandleConnection(conn); err != nil {
			log.Printf("Can't process connection, %s", err)
		}
	}
}
