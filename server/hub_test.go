package server_test

import (
	"encoding/json"
	"log"
	"net"
	"testing"
	"time"

	"github.com/nicholas-sokolov/omega/server"
	"github.com/stretchr/testify/require"
)

func TestHub(t *testing.T) {
	l, err := net.Listen("tcp", "localhost:8000")
	require.NoError(t, err)

	go func() {
		c, err := net.Dial("tcp", "127.0.0.1:8000")
		require.NoError(t, err)

		s := "{\"user_id\": 1, \"friends\": [2, 3, 4]}"
		_, err = c.Write([]byte(s))
		require.NoError(t, err)
	}()

	defer l.Close()

	log.Print("Start listening")

	hub := server.NewHub()
	go hub.Run()

	conn, err := l.Accept()
	if err != nil {
		log.Printf("Can't accept the connection, %s", err)

		return
	}

	var client server.Client
	if err := json.NewDecoder(conn).Decode(&client); err != nil {
		require.NoError(t, err)
	}

	log.Printf("User #%d has connected to the server", client.UserID)

	hub.Register <- &server.Client{
		Conn: &conn,
		Hub:  hub,
		User: server.User{
			UserID:   client.UserID,
			Friends:  client.Friends,
			IsOnline: true,
		},
	}

	time.Sleep(2 * time.Second)

	_, ok := hub.Clients[1]
	require.True(t, ok)
}
