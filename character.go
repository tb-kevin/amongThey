package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/jakecoffman/cp"
)

const (
	screenWidth  = 240
	screenHeight = 240

	// Background spritesheet
	tileSize    = 16
	tileXNum    = 25 // the number of 16px columns in the image width
	frameOX     = 0
	frameOY     = 32
	frameWidth  = 32
	frameHeight = 32
	frameNum    = 8
)

var (
	runnerImage *ebiten.Image
)

type Game struct {
	keys         []ebiten.Key
	bgLayers     [][]int
	count        int
	bgCollisions []int
	player       *ebiten.Image
	playerPosX   float64
	playerPosY   float64
	space        *cp.Space
}

var g *Game

func init() {
	eimg, _, err := ebitenutil.NewImageFromFile("images/tiles--mailbox.png")
	fmt.Println(eimg)
	if err != nil {
		log.Fatal(err)
	}
	space := cp.NewSpace()
	space.Iterations = 1 // Default: 10

	g = &Game{
		space:      space,
		player:     eimg,
		playerPosX: 0,
		playerPosY: 0,
	}
}

func (g *Game) Update() error {
	g.count++
	g.space.Step(1.0 / float64(ebiten.MaxTPS()))
	g.keys = inpututil.AppendPressedKeys(g.keys[:0])
	for _, k := range g.keys {
		if k == ebiten.KeyRight || k == ebiten.KeyD {
			g.playerPosX += 3
		} else if k == ebiten.KeyLeft || k == ebiten.KeyA {
			g.playerPosX -= 3
		} else if k == ebiten.KeyUp || k == ebiten.KeyW {
			g.playerPosY -= 3
		} else if k == ebiten.KeyDown || k == ebiten.KeyS {
			g.playerPosY += 3
		}
	}

	return nil
}
func (g *Game) Draw(screen *ebiten.Image) {

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-float64(frameWidth)/2, -float64(frameHeight)/2)
	op.GeoM.Translate(screenWidth/2, screenHeight/2)

	const xNum = screenWidth / tileSize // 15
	for _, l := range g.bgLayers {
		for i, t := range l {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64((i%xNum)*tileSize), float64((i/xNum)*tileSize))

			sx := (t % tileXNum) * tileSize
			sy := (t / tileXNum) * tileSize
			if g.bgCollisions[i] == 1 {
				body := cp.NewStaticBody()
				body.SetPosition(cp.Vector{X: float64(sx), Y: float64(sy)})
				shape := cp.NewBox(body, tileSize, tileSize, 0)
				shape.SetElasticity(0)
				shape.SetFriction(1)
				g.space.AddBody(shape.Body())
				g.space.AddShape(shape)
			}

		}
	}

	op.GeoM.Translate(float64(g.playerPosX), float64(g.playerPosY))

	body := cp.NewBody(1.0, cp.INFINITY)

	shape := cp.NewCircle(body, tileSize/2, cp.Vector{})
	shape.SetFriction(0)
	shape.SetElasticity(0)
	shape.SetCollisionType(1)
	body.SetPosition(cp.Vector{X: float64(g.playerPosX), Y: float64(g.playerPosY)})
	screen.DrawImage(g.player, op)

	ebitenutil.DebugPrint(screen,
		fmt.Sprintf(
			"TPS: %0.2f\n"+
				"PlayerX: %f\n"+
				"PlayerY: %f",
			ebiten.CurrentTPS(),
			g.playerPosX,
			g.playerPosY,
		),
	)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {

	img, _, err := image.Decode(bytes.NewReader(images.Runner_png))
	if err != nil {
		log.Fatal(err)
	}
	runnerImage = ebiten.NewImageFromImage(img)

	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("character collision")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
