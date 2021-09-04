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

var addr = flag.String("addr", ":8080", "http service address")
var board *rpc.Board
var lock sync.Mutex

var p1 = false
var p2 = false
var ready = make(chan bool)

var upgrader = websocket.Upgrader{} // use default options

type wsMessage struct {
	MT      int
	Message []byte
}

func processCommand(player int, message []byte) {
	lock.Lock()
	defer lock.Unlock()
	var command rpc.Command
	err := json.Unmarshal(message, &command)
	if err != nil {
		log.Println("error unmarshalling command:", err)
		return
	}

	board.Transfer(
		player,
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

func play1(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	commands := make(chan wsMessage)
	p1 = true

	<-ready

	done := make(chan bool)

	go func() {
		for {
			fmt.Println("waiting to read msg")
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				close(done)
				continue
			}
			commands <- wsMessage{mt, message}
		}
	}()
	t := time.NewTicker(10 * time.Millisecond)
	for {
		select {
		case wsMsg := <-commands:
			processCommand(1, wsMsg.Message)
			log.Printf("recv: %s", wsMsg.Message)
		case <-done:
			return
		case <-t.C:
			err = c.WriteMessage(1, boardToJSON())
			if err != nil {
				log.Println("write:", err)
			}
		}
	}
}

func play2(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	commands := make(chan wsMessage)
	p2 = true

	<-ready

	done := make(chan bool)

	go func() {
		for {
			fmt.Println("waiting to read msg")
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				close(done)
				break
			}
			commands <- wsMessage{mt, message}
		}
	}()
	t := time.NewTicker(30 * time.Millisecond)

	for {
		select {
		case wsMsg := <-commands:
			processCommand(2, wsMsg.Message)
			log.Printf("recv: %s", wsMsg.Message)
		case <-done:
			return
		case <-t.C:
			err = c.WriteMessage(1, boardToJSON())
			if err != nil {
				log.Println("write:", err)
			}
		}
	}
}

func main() {
	board = rpc.NewBoard(16)

	go func() {
		for {
			if p1 && p2 {
				board.Start()
				close(ready)
				return
			}
		}
	}()

	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/play/1", play1)
	http.HandleFunc("/play/2", play2)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
