package server

import (
	"encoding/json"
	"io"
	"log"
)

type Hub struct {
	// Registered Clients.
	Clients map[int]*Client

	// Register requests from the Clients.
	Register chan *Client

	// Unregister requests from Clients.
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[int]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			_, ok := h.Clients[client.UserID]
			// If this is a new client then need to broadcast to the each friend about his online status.
			if !ok {
				for _, friendID := range client.Friends {
					friend, ok := h.Clients[friendID]
					if !ok {
						continue
					}

					h.sendStatus(*friend.Conn, client.User)
					h.sendStatus(*client.Conn, friend.User)
				}
			} else {
				log.Printf("Attempt to connect od client #%d which already connected", client.UserID)
				conn := *client.Conn
				conn.Close()
				continue
			}

			h.Clients[client.UserID] = client

			go client.KeepConnection()
		case client := <-h.Unregister:
			c, ok := h.Clients[client.UserID]
			if !ok {
				continue
			}

			h.deleteClient(c)

			for _, f := range client.Friends {
				friend, ok := h.Clients[f]
				if !ok {
					continue
				}

				h.sendStatus(*friend.Conn, client.User)
			}
		}
	}
}

func (h Hub) sendStatus(w io.Writer, client User) {
	if err := json.NewEncoder(w).Encode(client); err != nil {
		log.Printf("Can't encode %s", err)
	}
}

func (h *Hub) deleteClient(client *Client) {
	conn := *client.Conn
	if err := conn.Close(); err != nil {
		log.Printf("Can't close the connection, %s", err)
	}

	delete(h.Clients, client.UserID)
}
