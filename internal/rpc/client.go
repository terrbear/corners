package rpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Client struct {
	boardUpdates chan Board
	clientID     string
	log          *log.Entry
	commands     chan Command
	playerID     PlayerID
}

func NewClient(playerID PlayerID, boardUpdates chan Board) *Client {
	clientID := uuid.New().String()
	client := &Client{
		clientID:     clientID,
		log:          log.WithField("clientID", clientID),
		commands:     make(chan Command),
		boardUpdates: boardUpdates,
		playerID:     playerID,
	}
	go client.start()

	return client
}

func (r *Client) SendCommand(command Command) {
	r.log.Debug("sending cmd to channel", command)
	r.commands <- command
}

func (r *Client) listen(ctx context.Context, c *websocket.Conn) error {
	errCount := 0
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			errCount++
			if errCount > 10 {
				return fmt.Errorf("error reading message: %s", err)
			}
			continue
		}

		board, err := DeserializeBoard(message)
		if err != nil {
			log.WithError(err).Error("error unmarshaling board")
			continue
		}
		log.Trace("board: ", board)
		r.boardUpdates <- *board
	}
}

func (r *Client) talk(ctx context.Context, c *websocket.Conn) error {
	r.log.Debug("waiting to talk to server...")
	for {
		select {
		case <-ctx.Done():
			log.Debug("ctx done")
			return nil
		case cmd := <-r.commands:
			log.Debug("sending command: ", cmd)
			msg, err := SerializeCommand(&cmd)
			if err != nil {
				log.Println("couldn't marshal command: ", err)
				continue
			}
			log.Trace("sending message: ", string(msg))
			err = c.WriteMessage(websocket.BinaryMessage, msg)
			if err != nil {
				log.Println("couldn't write command: ", err)
				continue
			}
		}
	}
}

func (r *Client) run() {
	u := ServerURL(r.playerID)
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error { return r.listen(ctx, c) })
	g.Go(func() error { return r.talk(ctx, c) })

	g.Wait()
}

func (r *Client) start() {
	for {
		r.run()
	}
}
