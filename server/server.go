package server

import (
	"encoding/json"
	"log"
	"net"
)

func RunServer() {
	l, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Printf("Can't listen the address, %s", err)

		return
	}
	defer l.Close()
	log.Print("Start listening")

	hub := NewHub()
	go hub.Run()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Can't accept the connection, %s", err)

			return
		}

		var client Client
		if err := json.NewDecoder(conn).Decode(&client); err != nil {
			log.Printf("Error of decoding, %s", err)

			return
		}

		log.Printf("User #%d has connected to the server", client.UserID)

		hub.Register <- &Client{
			Conn: &conn,
			Hub:  hub,
			User: User{
				UserID:   client.UserID,
				Friends:  client.Friends,
				IsOnline: true,
			},
		}
	}
}
