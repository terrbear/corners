package main

import (
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

const (
	maxPlayers = 4
)

var lock sync.Mutex

var upgrader = websocket.Upgrader{} // use default options

type wsMessage struct {
	MT      int
	Message []byte
}

func (gc *gameChannel) processCommand(player rpc.PlayerID, message []byte) {
	command, err := rpc.DeserializeCommand(message)
	if err != nil {
		log.Println("error unmarshalling command:", err)
		return
	}

	size := len(gc.board.Tiles)
	if command.SelectedX < 0 || command.SelectedX >= size ||
		command.SelectedY < 0 || command.SelectedY >= size {
		log.Warn("bad command given: ", command)
		return
	}

	gc.board.Transfer(
		player,
		gc.board.Tiles[command.SelectedX][command.SelectedY],
		gc.board.Tiles[command.TargetX][command.TargetY])
}

func (gc *gameChannel) serializedBoard(board *rpc.Board) []byte {
	b, err := rpc.SerializeBoard(board)
	if err != nil {
		log.WithError(err).Error("error marshaling board")
		return []byte{}
	}
	return b
}

var games = make(map[rpc.PlayerID]*gameChannel)

type gameChannel struct {
	players map[rpc.PlayerID]time.Time
	ready   chan bool
	board   *corners.Board
}

func NewGameChannel() *gameChannel {
	gc := gameChannel{
		players: make(map[rpc.PlayerID]time.Time),
		ready:   make(chan bool),
	}

	return &gc
}

// Obviously this is dirty; only call this if you're holding the lock
func startGame() {
	for p := range pendingGame.players {
		games[p] = pendingGame
	}
	log.Info("starting game!")
	players := make([]rpc.PlayerID, 0, len(pendingGame.players))
	for p := range pendingGame.players {
		players = append(players, p)
	}
	pendingGame.board = corners.NewBoard(env.Map(), players)
	err := pendingGame.board.Start()
	if err != nil {
		log.WithError(err).Error("error starting game")
	}

	close(pendingGame.ready)
	pendingGame = nil
}

var pendingGame *gameChannel

func timer() {
	ticker := time.NewTicker(env.LobbyTimeout())

	log.Info("lobby timeout: ", env.LobbyTimeout())

	for {
		lock.Lock()
		if pendingGame != nil {
			log.Debug("checking pending game; players len = ", len(pendingGame.players))
		}
		if pendingGame != nil && len(pendingGame.players) >= env.MinPlayers() {
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
		pendingGame.players[p] = time.Now()
	}

	if len(pendingGame.players) == maxPlayers {
		startGame()
		g := pendingGame
		pendingGame = nil
		return g
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

	var oldBoard *rpc.Board

	for {
		select {
		case wsMsg := <-commands:
			game.processCommand(id, wsMsg.Message)
			log.Tracef("recv: %s", wsMsg.Message)
		case <-done:
			return
		case <-t.C:
			board := game.board.ToRPCBoard(id)
			changes := board.Diff(oldBoard)
			oldBoard = board

			if len(changes.Tiles) > 0 || changes.Winner != nil {
				err = c.WriteMessage(websocket.BinaryMessage, game.serializedBoard(changes))
				if err != nil {
					log.Println("write:", err)
				}
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
