// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"terrbear.io/corners/internal/env"
	"terrbear.io/corners/internal/rpc"
	"terrbear.io/corners/server/corners"
)

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
	board := gc.board.ToRPCBoard()
	js, err := json.Marshal(board)
	if err != nil {
		log.WithError(err).Error("error marshaling board")
		return []byte{}
	}
	return js
}

var games = make(map[rpc.PlayerID]*gameChannel)

type gameChannel struct {
	players []rpc.PlayerID
	ready   chan bool
	board   *corners.Board
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
	log.Info("starting game!")
	pendingGame.board = corners.NewBoard(pendingGame.players, 16)
	pendingGame.board.Start()
	close(pendingGame.ready)
	pendingGame = nil
}

var pendingGame *gameChannel

const (
	maxPlayers = 4
	minPlayers = 1
)

func timer() {
	ticker := time.NewTicker(time.Second)

	for {
		lock.Lock()
		if pendingGame != nil {
			log.Debug("checking pending game; players len = ", len(pendingGame.players))
		}
		if pendingGame != nil && len(pendingGame.players) >= minPlayers {
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

	log.WithField("playerID", p).Debug("adding player")
	if len(pendingGame.players) < maxPlayers {
		pendingGame.players = append(pendingGame.players, p)
	}

	return pendingGame
}

// TODO cleanup games after they're done
// TODO clean up games that are expired

func play(w http.ResponseWriter, r *http.Request) {
	log.Trace("path: ", r.URL.Path)
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

	log.Debugf("player added to game with id %s; waiting for game to start\n", id)
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
	log.Info("starting server on port ", env.Port())
	go timer()
	http.HandleFunc("/play/", play)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", env.Port()), nil))
}
