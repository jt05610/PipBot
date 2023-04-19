/*
Copyright © 2023 Jonathan Taylor <jonrtaylor12@gmail.com>
*/

package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"go.bug.st/serial"
	"math"
	"strings"
	"sync/atomic"
	"time"
)

const Port = "/dev/cu.usbserial-1120"
const Baud = 115200

type Position struct {
	X float32
	Y float32
	Z float32
}

func (p *Position) Cmd(rate ...float64) []byte {
	var fr float64
	if len(rate) > 1 {
		fr = rate[0]
	} else {
		fr = 1000
	}
	return []byte(fmt.Sprintf("G0 F%v X%v Y%v Z%v\n", fr, p.X, p.Y, p.Z))
}

type CellType uint8

const (
	Tip CellType = iota
	Stock
	Standard
	Unknown
)

// Mixture describes what is in a cell.
// Mixtures can be stocks, mixtures of mixtures, dilutions, etc.
type Mixture struct {
	Contents map[Cell]float32
}

// Cell is the fundamental discrete addressable unit in the system.
// A cell can be a pipette tip position, an individual well of a plate, etc.
type Cell struct {
	Kind  CellType
	Empty bool
	*Position
	Content *Mixture
}

// Matrix is an aggregate of Cells. This can be a well plate, pipette tip box,
// tube rack, etc.
type Matrix struct {
	Cells   [][]*Position
	Home    *Position
	Rows    int
	Columns int
	kind    CellType
}

// Grid describes how individual Matrix units are arranged on the build plate.
type Grid struct {
	Rows []Matrix
}

func NewMatrix(home *Position, rowSpace, colSpace float32, nRow,
	nCol int, kind CellType) *Matrix {
	m := &Matrix{
		Cells:   make([][]*Position, nRow),
		Home:    home,
		Rows:    nRow,
		Columns: nCol,
		kind:    kind,
	}
	for row := 0; row < nRow; row++ {
		m.Cells[row] = make([]*Position, nCol)
		for col := 0; col < nCol; col++ {
			m.Cells[row][col] = &Position{
				X: home.X + (float32(col) * colSpace),
				Y: home.Y + (float32(row) * rowSpace),
				Z: home.Z,
			}
		}
	}
	return m
}

type PipBot struct {
	curPos *Position
	client serial.Port
	busy   atomic.Bool
	rx     chan []byte
	tx     chan []byte
	rate   float64
}

func NewPipBot(port string, baud int) *PipBot {
	var err error
	ret := &PipBot{
		rx: make(chan []byte),
		tx: make(chan []byte),
	}

	ret.client, err = serial.Open(port, &serial.Mode{
		BaudRate: baud,
	})
	if err != nil {
		panic(err)
	}
	return ret
}

func (b *PipBot) Close() {
	_ = b.client.Close()
}

func (b *PipBot) Home() {
	m := []byte("M302 P\n")
	b.tx <- m
	rcv := <-b.rx
	fmt.Println(rcv)
	m = []byte("G28\n")
	b.tx <- m
	for wait := true; wait; {
		rcv = <-b.rx
		fmt.Println(string(rcv))
		if !strings.Contains(string(rcv), "processing") {
			wait = false
		}
	}

	fmt.Println(rcv)
	b.curPos = &Position{
		X: 0,
		Y: 0,
		Z: 0,
	}
}

func TimeEst(start, end *Position, rate float64) time.Duration {
	dX := float64(end.X - start.X)
	dY := float64(end.Y - start.Y)
	dZ := float64(end.Z - start.Z)
	tot := math.Sqrt(math.Pow(dX, 2) + math.Pow(dY, 2) + math.Pow(dZ, 2))
	timeUs := int64((tot / rate) * float64(1000000))
	return time.Duration(timeUs) * time.Microsecond
}

func (b *PipBot) GoTo(p *Position) {
	target := b.curPos
	target.X = p.X
	target.Y = p.Y
	b.tx <- target.Cmd(b.rate)
	for wait := true; wait; {
		rcv := <-b.rx
		fmt.Println(string(rcv))
		if !strings.Contains(string(rcv), "processing") {
			wait = false
		}
	}
	b.curPos = target
	target.Z = p.Z
	b.tx <- target.Cmd(b.rate)
	for wait := true; wait; {
		rcv := <-b.rx
		fmt.Println(string(rcv))
		if !strings.Contains(string(rcv), "processing") {
			wait = false
		}
	}
}

func (b *PipBot) Clear() {
	target := &Position{
		X: 0,
		Y: 0,
		Z: 85,
	}
	b.GoTo(target)
}

func (b *PipBot) Eject() {
	target := &Position{
		X: 10,
		Y: b.curPos.Y,
		Z: 157,
	}
	b.GoTo(target)
	target.X = 15
	target.Z = 85
	b.GoTo(target)
}

func (b *PipBot) Sender(ctx context.Context) {
	for {
		select {
		case msg := <-b.tx:
			_, err := b.client.Write(msg)
			if err != nil {
				panic(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (b *PipBot) Listen(ctx context.Context) chan<- []byte {
	b.rx = make(chan []byte)
	go func() {
		defer close(b.rx)
		for {
			select {
			case <-ctx.Done():
			default:
				buf := make([]byte, 1024)
				n, err := b.client.Read(buf)
				if err != nil {
					panic(err)
				}
				if n > 0 {
					b.rx <- buf[:n-1]
				}
				time.Sleep(time.Duration(100) * time.Millisecond)

			}
		}
	}()

	return b.rx
}

// homeCmd represents the home command
var homeCmd = &cobra.Command{
	Use:   "home",
	Short: "homes the bot",
	Long:  `Sends G28. Be wary of clearances and things hitting other things!!`,
	Run: func(cmd *cobra.Command, args []string) {
		bot := NewPipBot(Port, Baud)
		bot.rate = 500
		ctx := context.Background()
		go bot.Sender(ctx)
		tips := NewMatrix(&Position{
			X: 165,
			Y: 103.5,
			Z: 73.5,
		}, 173.5-165, 173.5-165, 12, 8,
			Tip,
		)
		_ = bot.Listen(ctx)
		bot.Home()
		bot.Clear()
		bot.GoTo(tips.Home)
		p := tips.Home
		p.Z = 142
		bot.GoTo(p)
		bot.Eject()
	},
}

func init() {
	rootCmd.AddCommand(homeCmd)
}
