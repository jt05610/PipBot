package pipbot

import (
	"bufio"
	"context"
	"fmt"
	"go.bug.st/serial"
	"math"
	"strings"
	"sync/atomic"
	"time"
)

type PipBot struct {
	Layout   *Layout
	Current  *Position
	client   serial.Port
	busy     atomic.Bool
	rx       chan []byte
	Rate     float64
	TipStart int
}

func (b *PipBot) Bytes() []byte {
	return <-b.rx
}

func NewPipBot(port string, baud int, firstTip int) *PipBot {
	var err error
	ret := &PipBot{
		rx:       make(chan []byte),
		Layout:   MakeGrid(),
		TipStart: firstTip,
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
	m := []byte("G28\n")
	_, err := b.client.Write(m)
	if err != nil {
		panic(err)
	}
	buf := make([]byte, 1024)
	for wait := true; wait; {
		n, err := b.client.Read(buf)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(buf[:n]))
		if !strings.Contains(string(buf[:n]), "processing") {
			wait = false
		}
	}
	b.Current = &Position{
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
	target := b.Current
	target.X = p.X
	target.Y = p.Y
	_, err := b.client.Write(target.XY(b.Rate))
	if err != nil {
		panic(err)
	}
	target.Z = p.Z
	_, err = b.client.Write(target.Low(b.Rate))
	if err != nil {
		panic(err)
	}
	b.Current = target
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
		Y: b.Current.Y,
		Z: 156,
	}
	b.GoTo(target)
	target.Z = 85
	b.GoTo(target)
}

func (b *PipBot) Listen(ctx context.Context) bool {
	b.rx = make(chan []byte)
	scan := bufio.NewScanner(b.client)
	cont := true
	go func() {
		defer close(b.rx)
		for scan.Scan() {
			select {
			case <-ctx.Done():
				cont = false
			default:
				b.rx <- scan.Bytes()
				time.Sleep(time.Duration(100) * time.Millisecond)
			}
		}
	}()
	return cont
}
