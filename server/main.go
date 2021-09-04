// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"terrbear.io/corners/internal/rpc"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var board *rpc.Board
var lock sync.Mutex

var upgrader = websocket.Upgrader{} // use default options

type wsMessage struct {
	MT      int
	Message []byte
}

func processCommand(message []byte) {
	lock.Lock()
	defer lock.Unlock()
	var command rpc.Command
	err := json.Unmarshal(message, &command)
	if err != nil {
		log.Println("error unmarshalling command:", err)
		return
	}

	board.Transfer(
		board.Tiles[command.SelectedX][command.SelectedY],
		board.Tiles[command.TargetX][command.TargetY])
}

func boardToJSON() []byte {
	js, err := json.Marshal(board)
	if err != nil {
		log.Println("error marshaling board: ", err)
		return []byte{}
	}
	return js
}

func play(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	commands := make(chan wsMessage)
	go func() {
		for {
			fmt.Println("waiting to read msg")
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				continue
			}
			commands <- wsMessage{mt, message}
		}
	}()
	t := time.NewTicker(time.Second)
	for {
		select {
		case wsMsg := <-commands:
			processCommand(wsMsg.Message)
			log.Printf("recv: %s", wsMsg.Message)
		case <-t.C:
			err = c.WriteMessage(1, boardToJSON())
			if err != nil {
				log.Println("write:", err)
			}
		}
	}
}

func main() {
	board = rpc.NewBoard(8)

	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/play", play)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
