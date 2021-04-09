package server

import (
	"log"
	"net"
)

type User struct {
	UserID   int   `json:"user_id"`
	Friends  []int `json:"friends"`
	IsOnline bool  `json:"online"`
}

type Client struct {
	Conn *net.Conn
	Hub  *Hub
	User
}

func (c *Client) KeepConnection() {
	defer func() {
		c.IsOnline = false
		c.Hub.Unregister <- c
		log.Printf("User #%d has disconnected", c.UserID)
	}()

	for {
		conn := *c.Conn
		one := make([]byte, 1)

		_, err := conn.Read(one)
		if err != nil {
			return
		}
	}
}
