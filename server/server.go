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
	DuplexMode SendMode = iota
	ToFriend
	ForMyself
)

func (h *Server) HandleConnection(conn net.Conn) error {
	var user User
	if err := json.NewDecoder(conn).Decode(&user); err != nil {
		return err
	}

	sendStatus := func(user User, mode SendMode) {
		h.mu.Lock()
		defer h.mu.Unlock()

		for _, friendID := range user.Friends {
			friend, online := h.store[friendID]
			if !online {
				continue
			}

			switch mode {
			case DuplexMode:
				json.NewEncoder(friend.conn).Encode(user)
				json.NewEncoder(conn).Encode(friend)
			case ForMyself:
				json.NewEncoder(conn).Encode(friend)
			case ToFriend:
				json.NewEncoder(friend.conn).Encode(user)
			}
		}
	}

	// Check connection and put the new connection to the store.
	h.mu.Lock()
	oldConn, ok := h.store[user.UserID]

	mode := DuplexMode
	// If user's connection exists then broke the old connection.
	if ok {
		oldConn.conn.SetDeadline(time.Now())
		mode = ForMyself
	} else {
		log.Printf("User #%d connected", user.UserID)
		user.IsOnline = true
	}

	go sendStatus(user, mode)

	h.store[user.UserID] = &Conn{
		conn: conn,
		User: user,
	}
	h.mu.Unlock()

	//Run goroutine with block of reading from the connection
	//
	//When reading will be failed:
	//
	//os.ErrDeadlineExceeded  - nothing;
	//other reasons           - send message about it user's friends.
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
				sendStatus(user, ToFriend)

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
