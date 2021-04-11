package server

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

type Server struct {
	mu    sync.Mutex
	store map[int]*Conn // Kind of store of connections.
}

func NewServer() *Server {
	return &Server{
		store: make(map[int]*Conn),
	}
}

type Conn struct {
	conn net.Conn
	User
}

type User struct {
	UserID   int   `json:"user_id"`
	Friends  []int `json:"friends"`
	IsOnline bool  `json:"online"`
}

type SendMode int

const (
	Me SendMode = iota
	Them
	Duplex
)

func (h *Server) HandleConnection(conn net.Conn) error {
	var user User
	if err := json.NewDecoder(conn).Decode(&user); err != nil {
		return err
	}

	// Put new connection to the store.
	go func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		oldConn, ok := h.store[user.UserID]

		mode := Duplex
		// If user's connection exists then broke the old connection.
		if ok {
			oldConn.conn.SetDeadline(time.Now())
			mode = Me
		} else {
			log.Printf("User #%d connected", user.UserID)
		}

		user.IsOnline = true
		go h.sendStatus(conn, user, mode)

		h.store[user.UserID] = &Conn{
			conn: conn,
			User: user,
		}
	}()

	// Run goroutine with <-: connection.
	//
	// When reading will be failed:
	//
	// os.ErrDeadlineExceeded  - nothing;
	// other reasons           - send message about it user's friends.
	go func() {
		defer conn.Close()

		for {
			one := make([]byte, 1)
			_, err := conn.Read(one)
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					return
				}

				user.IsOnline = false
				h.sendStatus(conn, user, Them)

				h.mu.Lock()
				delete(h.store, user.UserID)
				h.mu.Unlock()

				log.Printf("User #%d disconnected", user.UserID)
				return
			}
		}
	}()

	return nil
}

func (h *Server) sendStatus(conn net.Conn, user User, mode SendMode) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, friendID := range user.Friends {
		friend, online := h.store[friendID]
		if !online {
			continue
		}

		switch mode {
		case Me:
			json.NewEncoder(conn).Encode(friend)
		case Them:
			json.NewEncoder(friend.conn).Encode(user)
		case Duplex:
			json.NewEncoder(friend.conn).Encode(user)
			json.NewEncoder(conn).Encode(friend)
		}
	}
}
