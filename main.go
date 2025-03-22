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

type Brick struct {
	Object
}

type Game struct {
	paddle    Paddle
	ball      Ball
	brick     Brick
	score     int
	highScore int
}

func main() {
	ebiten.SetWindowTitle("Pong in Ebitengine")
	ebiten.SetWindowSize(screenWidth, screenHeight)

	paddle := Paddle{
		Object: Object{
			X: screenWidth * 0.5,
			Y: screenHeight - 15,
			W: 90,
			H: 15,
		},
	}
	ball := Ball{
		Object: Object{
			X: screenWidth * 0.5,
			Y: screenHeight * 0.8,
			W: 15,
			H: 15,
		},
		dxdt: ballSpeed,
		dydt: ballSpeed,
	}

	brick := Brick{
		Object: Object{
			X: screenWidth * 0.5,
			Y: screenHeight * 0.1,
			W: 80,
			H: 25,
		},
	}

	g := &Game{
		paddle: paddle,
		ball:   ball,
		brick:  brick,
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
	vector.DrawFilledRect(screen,
		float32(g.brick.X), float32(g.brick.Y),
		float32(g.brick.W), float32(g.brick.H),
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
	g.CollideWithBrick()
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

func (g *Game) CollideWithBrick() {
	// Define ball and brick boundaries
	ballLeft := g.ball.X
	ballRight := g.ball.X + g.ball.W
	ballTop := g.ball.Y
	ballBottom := g.ball.Y + g.ball.H

	brickLeft := g.brick.X
	brickRight := g.brick.X + g.brick.W
	brickTop := g.brick.Y
	brickBottom := g.brick.Y + g.brick.H

	if ballRight > brickLeft && ballLeft < brickRight && ballBottom > brickTop && ballTop < brickBottom {
		//Find overlap on each side
		overlapLeft := ballRight - brickLeft
		overlapRight := brickRight - ballLeft
		overlapTop := ballBottom - brickTop
		overlapBot := brickBottom - ballTop

		//Locates minimum overlap
		minOverlap := overlapLeft
		collisionSide := "left"

		if overlapRight < minOverlap {
			minOverlap = overlapRight
			collisionSide = "right"
		}
		if overlapTop < minOverlap {
			minOverlap = overlapTop
			collisionSide = "top"
		}
		if overlapBot < minOverlap {
			minOverlap = overlapBot
			collisionSide = "bottom"
		}

		//reverse velocity based on collision side
		switch collisionSide {
		case "left", "right":
			g.ball.dxdt = -g.ball.dxdt
		case "top", "bottom":
			g.ball.dydt = -g.ball.dydt
		}

		//update score
		g.score++
		if g.score > g.highScore {
			g.highScore = g.score
		}
	}
}
