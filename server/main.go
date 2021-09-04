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
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"terrbear.io/corners/internal/rpc"
)

var addr = flag.String("addr", ":8080", "http service address")
var lock sync.Mutex

var ready = make(chan bool)

var upgrader = websocket.Upgrader{} // use default options

type wsMessage struct {
	MT      int
	Message []byte
}

func (gc *gameChannel) processCommand(player rpc.PlayerID, message []byte) {
	var command rpc.Command
	err := json.Unmarshal(message, &command)
	if err != nil {
		log.Println("error unmarshalling command:", err)
		return
	}

	gc.board.Transfer(
		player,
		gc.board.Tiles[command.SelectedX][command.SelectedY],
		gc.board.Tiles[command.TargetX][command.TargetY])
}

func (gc *gameChannel) boardToJSON() []byte {
	js, err := json.Marshal(gc.board)
	if err != nil {
		log.Println("error marshaling board: ", err)
		return []byte{}
	}
	return js
}

var games = make(map[rpc.PlayerID]*gameChannel)

type gameChannel struct {
	players []rpc.PlayerID
	ready   chan bool
	board   *rpc.Board
}

func NewGameChannel() *gameChannel {
	return &gameChannel{
		players: make([]rpc.PlayerID, 0),
		ready:   make(chan bool),
	}
}

var players = make(chan int)

// Obviously this is dirty; only call this if you're holding the lock
func startGame() {
	for _, p := range pendingGame.players {
		games[p] = pendingGame
	}
	fmt.Println("starting game!")
	pendingGame.board = rpc.NewBoard(pendingGame.players, 16)
	pendingGame.board.Start()
	close(pendingGame.ready)
	pendingGame = nil
}

var pendingGame *gameChannel

const (
	maxPlayers = 4
)

func timer() {
	ticker := time.NewTicker(20 * time.Second)

	for {
		lock.Lock()
		if pendingGame != nil {
			fmt.Println("checking pending game; players len = ", len(pendingGame.players))
		}
		if pendingGame != nil && len(pendingGame.players) >= 1 {
			startGame()
		}
		lock.Unlock()
		<-ticker.C
	}
}

func addPlayer(p rpc.PlayerID) *gameChannel {
	lock.Lock()
	defer lock.Unlock()

	if g, ok := games[p]; ok {
		return g
	}

	if pendingGame == nil {
		pendingGame = NewGameChannel()
	}

	fmt.Println("adding player: ", p)
	if len(pendingGame.players) < maxPlayers {
		pendingGame.players = append(pendingGame.players, p)
	}

	return pendingGame
}

// TODO cleanup games after they're done
// TODO clean up games that are expired

func play(w http.ResponseWriter, r *http.Request) {
	fmt.Println("path: ", r.URL.Path)
	id := rpc.PlayerID(strings.SplitAfter(r.URL.Path, "/play/")[1])

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	commands := make(chan wsMessage)

	done := make(chan bool)

	game := addPlayer(id)

	fmt.Printf("player added to game with id %s; waiting for game to start\n", id)
	<-game.ready

	go func() {
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				close(done)
				break
			}
			commands <- wsMessage{mt, message}
		}
	}()

	t := time.NewTicker(10 * time.Millisecond)

	for {
		select {
		case wsMsg := <-commands:
			game.processCommand(id, wsMsg.Message)
			log.Printf("recv: %s", wsMsg.Message)
		case <-done:
			return
		case <-t.C:
			err = c.WriteMessage(1, game.boardToJSON())
			if err != nil {
				log.Println("write:", err)
			}
		}
	}
}

func main() {
	go timer()
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/play/", play)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
