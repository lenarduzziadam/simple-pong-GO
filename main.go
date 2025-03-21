package main

import (
	"log"

	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const (
	screenWidth  = 650
	screenHeight = 450
	ballSpeed    = 3
	paddleSpeed  = 6
)

type Object struct {
	X, Y, W, H int
}

type Paddle struct {
	Object
}

type Ball struct {
	Object
	dxdt int
	dydt int
}

type Game struct {
	paddle    Paddle
	ball      Ball
	score     int
	highScore int
}

func main() {
	ebiten.SetWindowTitle("Pong in Ebitengine")
	ebiten.SetWindowSize(screenWidth, screenHeight)
	g := &Game{}
	err := ebiten.RunGame(g)
	if err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) Draw(screen *ebiten.Image) {
	vector.DrawFilledRect(screen,
		float32(g.paddle.X), float32(g.paddle.Y),
		float32(g.paddle.W), float32(g.paddle.H),
		color.White, false,
	)
	vector.DrawFilledRect(screen,
		float32(g.ball.X), float32(g.ball.Y),
		float32(g.ball.W), float32(g.ball.H),
		color.White, false,
	)
	scoreStr := "Score: " + fmt.Sprint(g.score)
	text.Draw(screen, scoreStr, basicfont.Face7x13, 10, 10, color.White)
}

func (g *Game) Update() error {
	return nil
}
