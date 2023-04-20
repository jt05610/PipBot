package pipbot

import (
	"context"
	"fmt"
	"go.bug.st/serial"
	"math"
	"strings"
	"sync/atomic"
	"time"
)

type PipBot struct {
	curPos *Position
	client serial.Port
	busy   atomic.Bool
	rx     chan []byte
	tx     chan []byte
	Rate   float64
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
	m := []byte("M302 P S0\n")
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
	b.tx <- target.Cmd(b.Rate)
	for wait := true; wait; {
		rcv := <-b.rx
		fmt.Println(string(rcv))
		if !strings.Contains(string(rcv), "processing") {
			wait = false
		}
	}
	b.curPos = target
	target.Z = p.Z
	b.tx <- target.Cmd(b.Rate)
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
