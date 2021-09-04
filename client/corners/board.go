package corners

import (
	"encoding/json"
	"fmt"
	"image"
	"log"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	"terrbear.io/corners/internal/rpc"
)

// Board represents the game board.
type Board struct {
	selectedX int
	selectedY int
	targetX   int
	targetY   int

	size   int
	tiles  [][]*Tile
	offset image.Point
	team   int
	init   sync.Once

	command chan rpc.Command
	board   rpc.Board
}

func (b *Board) runClient() {
	u := url.URL{Scheme: "ws", Host: "tannis.local:8080", Path: fmt.Sprintf("/play/%d", b.team)}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan bool)

	go func() {
		errCount := 0
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				errCount++
				if errCount > 10 {
					close(done)
					return
				}
				continue
			}

			var board rpc.Board
			err = json.Unmarshal(message, &board)
			if err != nil {
				log.Println("error unmarshaling board: ", err)
				continue
			}
			b.initBoard(board)
			b.board = board
		}
	}()

	for {
		select {
		case cmd := <-b.command:
			msg, err := json.Marshal(cmd)
			if err != nil {
				log.Println("couldn't marshal command: ", err)
				continue
			}
			fmt.Println("sending message: ", string(msg))
			err = c.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("couldn't write command: ", err)
				continue
			}
		case <-done:
			return
		}
	}
}

func (b *Board) startClient() {
	for {
		b.runClient()
	}
}

func (b *Board) initBoard(board rpc.Board) {
	b.init.Do(func() {
		fmt.Println("initializing board with size: ", len(board.Tiles))
		tiles := make([][]*Tile, len(board.Tiles))
		for x := range tiles {
			tiles[x] = make([]*Tile, len(board.Tiles))
			for y := range tiles[x] {
				tiles[x][y] = NewTile(&TileParams{x: x, y: y, resources: 1})
			}
		}

		b.tiles = tiles
		b.size = len(board.Tiles)
	})
}

// NewBoard generates a new Board with giving a size.
func NewBoard(player1 bool) *Board {
	b := &Board{
		command: make(chan rpc.Command),
	}

	if player1 {
		b.team = 1
	} else {
		b.team = 2
	}

	go b.startClient()

	return b
}

func (b *Board) translate(mouse *image.Point) (int, int) {
	if mouse == nil {
		return -1, -1
	}
	p := mouse.Sub(b.offset)
	if p.X < 0 || p.Y < 0 {
		return -1, -1
	}
	x := p.X / (tileSize + tileMargin)
	y := p.Y / (tileSize + tileMargin)
	return x, y
}

func (b *Board) forEach(x, y int, f func(int, int, *Tile) error) error {
	if len(b.board.Tiles) == 0 {
		return nil
	}
	for col := range b.tiles {
		for row := range b.tiles[col] {
			b.tiles[row][col].tile = b.board.Tiles[row][col]
			f(col, row, b.tiles[col][row])
		}
	}

	return nil
}

func (b *Board) transfer(sourceX, sourceY, targetX, targetY int) {
	b.command <- rpc.Command{
		SelectedX: sourceX,
		SelectedY: sourceY,
		TargetX:   targetX,
		TargetY:   targetY,
	}
}

// Update updates the board state.
func (b *Board) Update(input *Input) error {
	clickedX, clickedY := b.translate(input.LeftMouse())
	targetX, targetY := b.translate(input.RightMouse())

	if clickedX >= 0 && clickedY >= 0 {
		b.selectedX = clickedX
		b.selectedY = clickedY
		fmt.Printf("selected: %d, %d\n", clickedX, clickedY)
	}

	// fmt.Printf("selected: %+v\n", selected)
	if b.selectedX >= 0 && b.selectedY >= 0 && targetX >= 0 && targetY >= 0 {
		fmt.Printf("target: %d, %d\n", targetX, targetY)
		go func() {
			b.transfer(b.selectedX, b.selectedY, targetX, targetY)
		}()
	}

	return nil
}

// Size returns the board size.
func (b *Board) Size() (int, int) {
	x := b.size*tileSize + (b.size+1)*tileMargin
	y := x
	return x, y
}

// Draw draws the board to the given boardImage.
func (b *Board) Draw(boardImage *ebiten.Image) {
	boardImage.Fill(frameColor)

	for j := 0; j < b.size; j++ {
		for i := 0; i < b.size; i++ {
			v := 0
			op := &ebiten.DrawImageOptions{}
			x := i*tileSize + (i+1)*tileMargin
			y := j*tileSize + (j+1)*tileMargin
			op.GeoM.Translate(float64(x), float64(y))
			r, g, b, a := colorToScale(tileBackgroundColor(v))
			op.ColorM.Scale(r, g, b, a)
			/*if j == 0 && i == 0 {
				fmt.Printf("drawing tile bg image: %+v\n", tileImage.Bounds())
			}*/
			boardImage.DrawImage(tileImage, op)
		}
	}
	if len(b.tiles) > 0 {
		b.forEach(0, 0, func(x, y int, t *Tile) error {
			t.Draw(x, y, boardImage, &TileDrawParams{
				boardTeam: b.team,
				team:      t.team,
				targeted:  x == b.targetX && y == b.targetY,
				selected:  x == b.selectedX && y == b.selectedY,
			})
			return nil
		})
	}
}