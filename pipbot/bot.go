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
	Layout     *Layout
	Current    *Position
	TipChannel <-chan *Position
	client     serial.Port
	busy       atomic.Bool
	rx         chan []byte
	Rate       float64
	TipStart   int
	curTip     int
	cushion    float32
}

const (
	TipOffClear   float32 = 85
	TipOnClear    float32 = 142
	CushionVolume float32 = 25
)

func (b *PipBot) SetupDispenser() {
	m := []byte("M302 S1\n")

	_, err := b.client.Write(m)
	if err != nil {
		panic(err)
	}
	m = []byte("M82\n")

	_, err = b.client.Write(m)
	if err != nil {
		panic(err)
	}
	m = []byte("G92 E0\n")
	_, err = b.client.Write(m)
	if err != nil {
		panic(err)
	}
}

// Init gets ready to run a protocol. Note that it automatically selects the last matrix as the tip matrix -- this will
// not be hardcoded in less time pressed versions :)
func (b *PipBot) Init() {
	b.TipChannel = b.Layout.Matrices[3].Channel()
	b.cushion = CushionVolume
	for b.curTip != b.TipStart {
		b.curTip++
		_ = <-b.TipChannel
	}
	b.Home()
	b.SetupDispenser()
	target := b.Current
	target.Z = TipOffClear
	m := []byte("G92 E-30\n")
	_, err := b.client.Write(m)
	if err != nil {
		panic(err)
	}
	b.Dispense()
	b.ResetCush()
	m = []byte("G92 E0\n")
	_, err = b.client.Write(m)
	if err != nil {
		panic(err)
	}
	b.Do(target)
}

func (b *PipBot) Bytes() []byte {
	return <-b.rx
}

// getTip gets the next tip position and increments the counter
func (b *PipBot) getTip() *Position {
	b.curTip++
	return <-b.TipChannel
}

func (b *PipBot) Transfer(src *Cell, dest *Cell, vol float32, eject bool) {
	// get increment tip id and pickup the tip
	t := b.getTip()
	b.Do(t)
	t.Z = TipOnClear
	b.Do(t)

	// go to source and insert into fluid
	t = src.Position
	b.Do(t)

	// draw fluid
	b.Pickup(vol)

	// remove from container
	t.Z = TipOnClear
	b.Do(t)

	// go to dest and insert into fluid
	t = dest.Position
	b.Do(t)

	// dispense fluid
	b.Dispense()
	time.Sleep(time.Duration(1) * time.Second)
	// remove from container
	t.Z = TipOnClear
	b.Do(t)

	b.ResetCush()

	if eject {
		b.Eject()
	}
}

func NewPipBot(port string, baud int, firstTip int) *PipBot {
	var err error
	ret := &PipBot{
		rx:       make(chan []byte),
		Layout:   MakeGrid(),
		TipStart: firstTip,
		curTip:   0,
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

func (b *PipBot) Do(target *Position) {
	dX := float64(target.X - b.Current.X)
	dY := float64(target.Y - b.Current.Y)
	dZ := float64(target.Z - b.Current.Z)
	b.GoTo(target)
	if (dX != 0) || (dY != 0) {
		dP := math.Sqrt(math.Pow(dX, 2) + math.Pow(dY, 2))
		t := dP / b.Rate
		time.Sleep(time.Duration(t) * time.Second)
	}
	if dZ != 0 {
		t := math.Abs(dZ / 10)
		time.Sleep(time.Duration(t*1000000) * time.Microsecond)
	}
}

func (b *PipBot) Pickup(volume float32) {
	travel := volume / 10
	target := Position{Z: 85}
	target.Z = target.Z + 1
	m := []byte(fmt.Sprintf("G1 F500 E-%v\n", travel))
	_, err := b.client.Write(m)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Duration(500) * time.Millisecond)
}

func (b *PipBot) Dispense() {
	m := []byte("G1 F500 E0\n")
	_, err := b.client.Write(m)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Duration(500) * time.Millisecond)
}

func (b *PipBot) ResetCush() {
	m := []byte("G1 F500 E0\n")
	_, err := b.client.Write(m)
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Duration(500) * time.Millisecond)
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

func (b *PipBot) Eject() {
	target := &Position{
		X: 10,
		Y: b.Current.Y,
		Z: 156,
	}
	b.Do(target)
	target.Z = 85
	b.Do(target)
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
