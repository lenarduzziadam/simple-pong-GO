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
	screenWidth  = 1550
	screenHeight = 900
	ballSpeed    = 5
	paddleSpeed  = 8
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

	paddle := Paddle{
		Object: Object{
			X: screenWidth * 0.5,
			Y: screenHeight - 20,
			W: 90,
			H: 15,
		},
	}
	ball := Ball{
		Object: Object{
			X: 150,
			Y: 150,
			W: 15,
			H: 15,
		},
		dxdt: ballSpeed,
		dydt: ballSpeed,
	}

	g := &Game{
		paddle: paddle,
		ball:   ball,
	}
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

	highScoreStr := "High Score: " + fmt.Sprint(g.highScore)
	text.Draw(screen, highScoreStr, basicfont.Face7x13, 10, 30, color.White)
}

func (g *Game) Update() error {
	g.paddle.MoveOnKeyPress()
	g.ball.Move()
	g.CollideWithWall()
	g.CollideWithPaddle()
	return nil
}

func (p *Paddle) MoveOnKeyPress() {
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		p.X += paddleSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		p.X -= paddleSpeed
	}

}

func (b *Ball) Move() {
	b.X += b.dxdt
	b.Y += b.dydt
}

func (g *Game) Reset() {
	g.ball.X = 100
	g.ball.Y = 250

	g.score = 0
}

func (g *Game) CollideWithWall() {
	if g.ball.Y >= screenHeight {
		g.Reset()

	} else if g.ball.X <= 0 {
		g.ball.dxdt = ballSpeed

	} else if g.ball.Y <= 0 {
		g.ball.dydt = ballSpeed

	} else if g.ball.X >= screenWidth {
		g.ball.dxdt = -ballSpeed
	}
}

func (g *Game) CollideWithPaddle() {
	if g.ball.Y >= g.paddle.Y && g.ball.X >= g.paddle.X && g.ball.X <= g.paddle.X+g.paddle.W {
		g.ball.dydt = -g.ball.dydt
		g.score++
		if g.score > g.highScore {
			g.highScore = g.score
		}
	}
}
