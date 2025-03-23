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

var BrickColors = map[int]color.RGBA{
	4: {255, 0, 0, 255},     // Red
	3: {0, 0, 255, 255},     // Blue
	2: {0, 255, 0, 255},     // Green
	1: {255, 255, 255, 255}, // White
}

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

	Hit       bool
	Health    int
	MaxHealth int
}

type Game struct {
	paddle    Paddle
	ball      Ball
	bricks    []Brick
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

	g := &Game{
		paddle: paddle,
		ball:   ball,
	}
	g.initBricks()
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
	for _, brick := range g.bricks {
		if brick.Hit {
			continue
		}
		brickColor := BrickColors[brick.Health]
		vector.DrawFilledRect(screen,
			float32(brick.X), float32(brick.Y),
			float32(brick.W), float32(brick.H),
			brickColor, false,
		)
	}
	scoreStr := "Score: " + fmt.Sprint(g.score)
	text.Draw(screen, scoreStr, basicfont.Face7x13, 10, 10, color.White)

	highScoreStr := "High Score: " + fmt.Sprint(g.highScore)
	text.Draw(screen, highScoreStr, basicfont.Face7x13, 10, 30, color.White)
}

func (g *Game) Update() error {
	//updates methods in order (so call them in the order you want them on screen)
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
	g.ball.X = screenWidth * 0.4
	g.ball.Y = screenHeight * 0.7

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

// new Brick initializer field
func (g *Game) initBricks() {
	//initializes row amount, column amoutn, width height, and padding variable as well as a start y and x at pposition (20,20)
	rows := 10
	cols := 18
	brickWidth := 78
	brickHeight := 25
	padding := 5
	startX := 20
	startY := 20

	//creates array of Brick type struct
	g.bricks = []Brick{}

	//typical iteration logic rows, handles until final row reached
	for row := 0; row < rows; row++ {
		//innerloop goes over columns for each row (making an inner for and a time complexity of at least O(n^2)
		//this inner loop also initializes an initial x, y position keeping padding in mind and then pads out the bricks speerating them by 5 pixels in this case but still plenty capable to  figure it out
		for col := 0; col < cols; col++ {
			x := startX + col*(brickWidth+padding)
			y := startY + row*(brickHeight+padding)
			//assigns varibales to brick, and Bricks inner Object sturct
			brick := Brick{
				Object: Object{
					X: x,
					Y: y,
					W: brickWidth,
					H: brickHeight,
				},
				Hit:       false,
				Health:    4,
				MaxHealth: 4,
			}
			g.bricks = append(g.bricks, brick)
		}
	}
}
func (g *Game) CollideWithBrick() {
	// Define ball and brick boundaries
	ballLeft := g.ball.X
	ballRight := g.ball.X + g.ball.W
	ballTop := g.ball.Y
	ballBottom := g.ball.Y + g.ball.H

	//making for loop for iteration
	for i := range g.bricks {
		//now initializes a brick as the current i positone dbrick, gotten via iteration
		brick := &g.bricks[i]
		//checks if brick health is less of equal to 0 if so will remove the box from the screen, no collision or drawed rectangle anymore as far as i can guess but maybe be careful
		if brick.Health <= 0 {
			continue
		}

		//intiailized more easy to comprehend and mroe universal variables for many collisions/object situations can always edit later on. .

		brickLeft := brick.X
		brickRight := brick.X + brick.W
		brickTop := brick.Y
		brickBottom := brick.Y + brick.H

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
			brick.Health-- //Decrement brick health
			//update score
			g.score++
			if g.score > g.highScore {
				g.highScore = g.score
			}

			if brick.Health <= 0 {
				brick.Hit = true
				g.score += 10
				if g.score > g.highScore {
					g.highScore = g.score
				}
			}
			break
		}

	}
}
