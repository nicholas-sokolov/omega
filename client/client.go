package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net"

	"github.com/nicholas-sokolov/omega/server"
)

func main() {
	var fileName string

	flag.StringVar(
		&fileName,
		"file",
		"",
		"json file name where defined user and his friends like this {\"user_id\": 1, \"friends\": [2, 3, 4]}")
	flag.Parse()

	if len(fileName) == 0 {
		log.Fatal("file can't be empty")
	}

	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.Dial("tcp", "127.0.0.1:8000")
	if err != nil {
		log.Fatalf("Can't set the connection, %s", err)
	}

	_, err = conn.Write(b)
	if err != nil {
		log.Fatal(err)
	}

	for {
		var user server.User
		if err := json.NewDecoder(conn).Decode(&user); err != nil {
			log.Print("Server disconnected")

			break
		}

		if user.IsOnline {
			log.Printf("Friend #%d is online now", user.UserID)
		} else {
			log.Printf("Friend #%d has left", user.UserID)
		}
	}
}
