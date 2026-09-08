package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	memory     [4096]byte
	display    [32][64]bool
	PC         uint16
	I          uint16
	stack      []uint16
	delayTimer uint8
	soundTimer uint8
	V          [16]uint8
}

var font = [...]byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

const (
	fontStart = 0x050
	romStart  = 0x200
)

func (g *Game) Init(rom []byte) {
	copy(g.memory[fontStart:], font[:])
	copy(g.memory[romStart:], rom)
	g.PC = romStart
}
func (g *Game) step() {
	opcode := uint16(g.memory[g.PC])<<8 | uint16(g.memory[g.PC+1])
	g.PC += 2
	X := (opcode & 0x0F00) >> 8
	Y := (opcode & 0x00F0) >> 4
	N := opcode & 0x000F
	NN := opcode & 0x00FF
	NNN := opcode & 0x0FFF
	firstNible := (opcode & 0xF000) >> 12
	switch firstNible {
	case 0x0:
		switch opcode {
		case 0x00E0:
			g.display = [32][64]bool{}
		case 0x00EE:
			g.PC = g.stack[len(g.stack)-1]
			g.stack = g.stack[:len(g.stack)-1]
		}
	case 0x1:
		g.PC = NNN
	case 0x2:
		g.stack = append(g.stack, g.PC)
		g.PC = NNN
	case 0x3:
		if uint16(g.V[X]) == NN {
			g.PC += 2
		}
	case 0x4:
		if uint16(g.V[X]) != NN {
			g.PC += 2
		}
	case 0x5:
		if g.V[X] == g.V[Y] {
			g.PC += 2
		}
	case 0x9:
		if g.V[X] != g.V[Y] {
			g.PC += 2
		}
	case 0x6:
		g.V[X] = uint8(NN)
	case 0x7:
		g.V[X] += uint8(NN)
	case 0x8:
		switch N {
		case 0x0:

			g.V[X] = g.V[Y]
		case 0x1:
			g.V[X] = g.V[X] | g.V[Y]
		case 0x2:

			g.V[X] = g.V[X] & g.V[Y]
		case 0x3:
			g.V[X] = g.V[X] ^ g.V[Y]
		case 0x4:
			sum := uint16(g.V[X]) + uint16(g.V[Y])
			g.V[X] = uint8(sum)
			if sum > 255 {
				g.V[0xF] = 1
			} else {
				g.V[0xF] = 0
			}
		case 0x5:
			flag := uint8(0)
			if g.V[X] >= g.V[Y] {
				flag = 1
			}
			g.V[X] = g.V[X] - g.V[Y]
			g.V[0xF] = flag
		case 0x7:
			flag := uint8(0)
			if g.V[Y] >= g.V[X] {
				flag = 1
			}
			g.V[X] = g.V[Y] - g.V[X]
			g.V[0xF] = flag
		case 0x6:
			carry := g.V[X] & 1
			g.V[X] >>= 1
			g.V[0xF] = carry
		case 0xE:
			carry := (g.V[X] >> 7) & 1
			g.V[X] <<= 1
			g.V[0xF] = carry
		}
	case 0xA:
		g.I = NNN
	case 0xB:
		g.PC = NNN + uint16(g.V[0])
	case 0xC:
		r := rand.Uint32()
		g.V[X] = uint8(r) & uint8(NN)
	case 0xD:
		x := int(g.V[X] % 64)
		y := int(g.V[Y] % 32)
		g.V[0xF] = 0
		for row := 0; uint16(row) < N; row++ {
			b := g.memory[g.I+uint16(row)]
			if y+row >= 32 {
				break
			}
			for col := 0; col < 8; col++ {
				if x+col >= 64 {
					continue
				}
				pix := (b >> (7 - col)) & 1
				if pix == 0 {
					continue
				}
				curr := g.display[y+row][x+col]
				if curr {
					g.V[0xF] = 1
				}
				g.display[y+row][x+col] = !g.display[y+row][x+col]
			}
		}
	default:
		fmt.Printf("Unknown opcode: 0x%04X\n", opcode)
	}
}
func (g *Game) Update() error {
	for i := 0; i < 10; i++ {
		g.step()
	}
	return nil
}
func (g *Game) Draw(screen *ebiten.Image) {
	for row := 0; row < len(g.display); row++ {
		for col := 0; col < len(g.display[row]); col++ {
			if g.display[row][col] {
				screen.Set(col, row, color.RGBA{255, 51, 68, 1.0})
			} else {
				screen.Set(col, row, color.RGBA{102, 0, 0, 1.0})
			}
		}
	}
}
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 64, 32
}
func main() {
	ebiten.SetWindowSize(640, 320)
	ebiten.SetWindowTitle("CHIP-8")
	game := &Game{}
	romPath := "chip8-test-suite/bin/1-chip8-logo.ch8"
	rom, err := os.ReadFile(romPath)
	if err != nil {
		log.Fatalf("Failed to read file: %s", err)
	}
	game.Init(rom)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
