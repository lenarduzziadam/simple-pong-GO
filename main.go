package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const (
	screenWidth  = 1250
	screenHeight = 750
	ballSpeed    = 5
	paddleSpeed  = 8
)

//go:embed music/brickBreaker2.wav
var themeWAV []byte

//go:embed music/level1BB.wav
var level1WAV []byte

//go:embed music/level2BB.wav
var level2WAV []byte

//go:embed music/level3BB.wav
var level3WAV []byte

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
	paddle       Paddle
	ball         Ball
	bricks       []Brick
	score        int
	highScore    int
	currentLevel int
	mode         string //campaign or random
	gameState    string //title, playing, win, game-over
	ballLaunched bool
	audioContext *audio.Context
	themePlayer  *audio.Player
	currentTrack *audio.Player
	levelTracks  map[int][]byte
}

var levelTracks = map[int][]byte{
	0: level1WAV,
	1: level2WAV,
	2: level3WAV,
}

func main() {
	ebiten.SetWindowTitle("Adam's Brickbreak")
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

	audioContext := audio.NewContext(44100)

	stream, err := wav.DecodeWithSampleRate(audioContext.SampleRate(), bytes.NewReader(themeWAV))
	if err != nil {
		log.Fatal(err)
	}
	loop := audio.NewInfiniteLoop(stream, stream.Length())
	player, err := audio.NewPlayer(audioContext, loop)
	if err != nil {
		log.Fatal(err)
	}
	player.SetVolume(0.6)
	player.Play()
	fmt.Println("🎵 Theme music started!")

	g := &Game{
		paddle:       paddle,
		ball:         ball,
		gameState:    "title",
		mode:         "random",
		audioContext: audioContext,
		themePlayer:  player,
		levelTracks:  levelTracks,
	}

	err = ebiten.RunGame(g)
	if err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.gameState {
	case "title":
		text.Draw(screen, "BRICK BREAKER", basicfont.Face7x13, 600, 200, color.White)
		text.Draw(screen, "Press 1 for Campaign", basicfont.Face7x13, 600, 250, color.White)
		text.Draw(screen, "Press 2 for Random Mode", basicfont.Face7x13, 600, 270, color.White)
	case "playing":
		// Draw paddle, ball, bricks, score
		// (your existing draw code here)
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
}

func (g *Game) Update() error {

	switch g.gameState {
	case "title":
		if ebiten.IsKeyPressed(ebiten.Key1) {
			g.mode = "campaign"
			g.currentLevel = 0
			g.playLevelMusic(0)
			g.initBricks()
			g.gameState = "playing"
		} else if ebiten.IsKeyPressed(ebiten.Key2) {
			g.mode = "random"
			g.currentLevel = 0
			g.initBricks()
			g.gameState = "playing"
		}

	case "playing":
		g.paddle.MoveOnKeyPress()

		// If ball hasn't launched, follow paddle
		if !g.ballLaunched {
			g.ball.X = g.paddle.X + g.paddle.W/2 - g.ball.W/2
			g.ball.Y = g.paddle.Y - g.ball.H

			// Launch on spacebar press
			if ebiten.IsKeyPressed(ebiten.KeySpace) {
				g.ballLaunched = true
				g.ball.dxdt = ballSpeed
				g.ball.dydt = -ballSpeed
			}
		} else {
			// Ball moves normally
			g.ball.Move()
		}

		g.CollideWithWall()
		g.CollideWithPaddle()
		g.CollideWithBrick()

		//checks if all bricks allBricksCleared
		if g.allBricksCleared() {
			if g.mode == "campaign" {
				g.currentLevel++
				if g.currentLevel >= len(Levels) {
					// End of campaign — reset or go to title screen
					g.currentLevel = 0
					g.gameState = "title"
					g.score = 0
					g.Reset()
					return nil
				}
			}
			// For both modes: regenerate bricks and reset ball
			g.playLevelMusic(g.currentLevel)
			g.initBricks()
			tempScore := g.score
			g.Reset()
			g.score += tempScore
		}

	}
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
	g.score = 0
	g.ball.X = g.paddle.X + g.paddle.W/2 - g.ball.W/2
	g.ball.Y = g.paddle.Y - g.ball.H
	g.ball.dxdt = 0
	g.ball.dydt = 0
	g.ballLaunched = false
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
func (g *Game) allBricksCleared() bool {
	for _, b := range g.bricks {
		if b.Health > 0 {
			return false
		}
	}
	return true
}

// new Brick initializer field
func (g *Game) initBricks() {
	g.bricks = []Brick{}

	// Shared padding & offset
	padding := 10
	startX := padding
	startY := padding * 2 // leaves some space from the top

	// Setup for both modes
	var rows, cols int

	if g.mode == "campaign" {
		layout := Levels[g.currentLevel]
		rows = len(layout)
		if rows > 0 {
			cols = len(layout[0])
		}
	} else if g.mode == "random" {
		rows = rand.Intn(6) + 6  // 6–11 rows
		cols = rand.Intn(8) + 10 // 10–17 columns
	}

	// ✅ Calculate scalable dimensions
	totalPaddingX := (cols + 1) * padding
	totalPaddingY := (rows + 1) * padding

	brickWidth := (screenWidth - totalPaddingX) / cols
	brickHeight := ((screenHeight / 2) - totalPaddingY) / rows // fit in top half

	if g.mode == "campaign" {
		layout := Levels[g.currentLevel]
		for row := 0; row < len(layout); row++ {
			for col := 0; col < len(layout[row]); col++ {
				health := layout[row][col]
				x := startX + col*(brickWidth+padding)
				y := startY + row*(brickHeight+padding)

				brick := Brick{
					Object: Object{
						X: x,
						Y: y,
						W: brickWidth,
						H: brickHeight,
					},
					Hit:       false,
					Health:    health,
					MaxHealth: health,
				}
				g.bricks = append(g.bricks, brick)
			}
		}
	} else if g.mode == "random" {
		for row := 0; row < rows; row++ {
			for col := 0; col < cols; col++ {
				if rand.Float64() < 0.5 {
					continue
				}
				health := rand.Intn(4) + 1
				x := startX + col*(brickWidth+padding)
				y := startY + row*(brickHeight+padding)

				brick := Brick{
					Object: Object{
						X: x,
						Y: y,
						W: brickWidth,
						H: brickHeight,
					},
					Hit:       false,
					Health:    health,
					MaxHealth: health,
				}
				g.bricks = append(g.bricks, brick)
			}
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

func (g *Game) playLevelMusic(level int) {
	// Stop old track if it's playing
	if g.currentTrack != nil && g.currentTrack.IsPlaying() {
		g.currentTrack.Pause()
	}

	trackData, ok := g.levelTracks[level]
	if !ok {
		trackData = themeWAV // fallback/default
	}

	stream, err := wav.DecodeWithSampleRate(g.audioContext.SampleRate(), bytes.NewReader(trackData))
	if err != nil {
		log.Println("Failed to decode level music:", err)
		return
	}

	loop := audio.NewInfiniteLoop(stream, stream.Length())
	player, err := audio.NewPlayer(g.audioContext, loop)
	if err != nil {
		log.Println("Failed to create audio player:", err)
		return
	}

	player.SetVolume(0.5)
	player.Play()
	g.currentTrack = player
}
