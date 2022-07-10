package main

import (
	"fmt"
	"image"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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
var spinner = []byte(`-\|/`)

// Game is an isometric demo game.
type Game struct {
	w, h         int
	currentLevel *Level

	camX, camY float64
	camScale   float64
	camScaleTo float64

	mousePanX, mousePanY int

	spinnerIndex int

	keys         []ebiten.Key
	bgLayers     [][]int
	count        int
	bgCollisions []int
	player       *ebiten.Image
	playerPosX   float64
	playerPosY   float64
	space        *cp.Space
}

// NewGame returns a new isometric demo Game.
func NewGame() (*Game, error) {
	l, err := NewLevel()
	if err != nil {
		return nil, fmt.Errorf("failed to create new level: %s", err)
	}
	eimg, _, err := ebitenutil.NewImageFromFile("images/tiles--mailbox.png")
	fmt.Println(eimg)
	if err != nil {
		log.Fatal(err)
	}
	space := cp.NewSpace()
	space.Iterations = 1 // Default: 10

	g := &Game{
		currentLevel: l,
		camScale:     2,
		camScaleTo:   2,
		mousePanX:    math.MinInt32,
		mousePanY:    math.MinInt32,
		space:        space,
		player:       eimg,
		playerPosX:   0,
		playerPosY:   0,
	}
	return g, nil
}

// Update reads current user input and updates the Game state.
func (g *Game) Update() error {
	// Update target zoom level.
	var scrollY float64
	if ebiten.IsKeyPressed(ebiten.KeyC) || ebiten.IsKeyPressed(ebiten.KeyPageDown) {
		scrollY = -0.25
	} else if ebiten.IsKeyPressed(ebiten.KeyE) || ebiten.IsKeyPressed(ebiten.KeyPageUp) {
		scrollY = .25
	} else {
		_, scrollY = ebiten.Wheel()
		if scrollY < -1 {
			scrollY = -1
		} else if scrollY > 1 {
			scrollY = 1
		}
	}
	g.camScaleTo += scrollY * (g.camScaleTo / 7)

	// Clamp target zoom level.
	if g.camScaleTo < 0.01 {
		g.camScaleTo = 0.01
	} else if g.camScaleTo > 100 {
		g.camScaleTo = 100
	}

	// Smooth zoom transition.
	div := 10.0
	if g.camScaleTo > g.camScale {
		g.camScale += (g.camScaleTo - g.camScale) / div
	} else if g.camScaleTo < g.camScale {
		g.camScale -= (g.camScale - g.camScaleTo) / div
	}

	// Pan camera via keyboard.
	pan := 7.0 / g.camScale
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.camX -= pan
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.camX += pan
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		g.camY -= pan
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.camY += pan
	}
	// Pan camera via mouse.
	// if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
	// 	if g.mousePanX == math.MinInt32 && g.mousePanY == math.MinInt32 {
	// 		g.mousePanX, g.mousePanY = ebiten.CursorPosition()
	// 	} else {
	// 		x, y := ebiten.CursorPosition()
	// 		dx, dy := float64(g.mousePanX-x)*(pan/100), float64(g.mousePanY-y)*(pan/100)
	// 		g.camX, g.camY = g.camX-dx, g.camY+dy
	// 	}
	// } else if g.mousePanX != math.MinInt32 || g.mousePanY != math.MinInt32 {
	// 	g.mousePanX, g.mousePanY = math.MinInt32, math.MinInt32
	// }

	// Clamp camera position.
	worldWidth := float64(g.currentLevel.w * g.currentLevel.tileSize / 2)
	worldHeight := float64(g.currentLevel.h * g.currentLevel.tileSize / 2)
	if g.camX < worldWidth*-1 {
		g.camX = worldWidth * -1
	} else if g.camX > worldWidth {
		g.camX = worldWidth
	}
	if g.camY < worldHeight*-1 {
		g.camY = worldHeight * -1
	} else if g.camY > 0 {
		g.camY = 0
	}

	// Randomize level.
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		l, err := NewLevel()
		if err != nil {
			return fmt.Errorf("failed to create new level: %s", err)
		}

		g.currentLevel = l
	}
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

// Draw draws the Game on the screen.
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
	// Render level.
	g.renderLevel(screen)

	// Print game info.
	debugBox := image.NewRGBA(image.Rect(0, 0, g.w, 200))
	debugImg := ebiten.NewImageFromImage(debugBox)
	// ebitenutil.DebugPrint(debugImg, fmt.Sprintf("KEYS WASD EC R\nFPS  %0.0f\nTPS  %0.0f\nSCA  %0.2f\nPOS  %0.0f,%0.0f", ebiten.CurrentFPS(), ebiten.CurrentTPS(), g.camScale, g.camX, g.camY))
	// op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(3, 0)
	op.GeoM.Scale(2, 2)
	screen.DrawImage(debugImg, op)
}

// Layout is called when the Game's layout changes.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	s := ebiten.DeviceScaleFactor()
	g.w, g.h = int(s*float64(outsideWidth)), int(s*float64(outsideHeight))
	return g.w, g.h
}

// cartesianToIso transforms cartesian coordinates into isometric coordinates.
func (g *Game) cartesianToIso(x, y float64) (float64, float64) {
	tileSize := g.currentLevel.tileSize
	ix := (x - y) * float64(tileSize/2)
	iy := (x + y) * float64(tileSize/4)
	return ix, iy
}

/*
// isoToCartesian transforms isometric coordinates into cartesian coordinates.
func (g *Game) isoToCartesian(x, y float64) (float64, float64) {
	tileSize := g.currentLevel.tileSize
	cx := (x/float64(tileSize/2) + y/float64(tileSize/4)) / 2
	cy := (y/float64(tileSize/4) - (x / float64(tileSize/2))) / 2
	return cx, cy
}
*/

// renderLevel draws the current Level on the screen.
func (g *Game) renderLevel(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	var t *Tile
	for y := 0; y < g.currentLevel.h; y++ {
		for x := 0; x < g.currentLevel.w; x++ {
			xi, yi := g.cartesianToIso(float64(x), float64(y))

			// Skip drawing off-screen tiles.
			padding := float64(g.currentLevel.tileSize) * g.camScale
			drawX, drawY := ((xi-g.camX)*g.camScale)+float64(g.w/2.0), ((yi+g.camY)*g.camScale)+float64(g.h/2.0)
			if drawX+padding < 0 || drawY+padding < 0 || drawX > float64(g.w) || drawY > float64(g.h) {
				continue
			}

			t = g.currentLevel.tiles[y][x]
			if t == nil {
				continue // No tile at this position.
			}

			op.GeoM.Reset()
			// Move to current isometric position.
			op.GeoM.Translate(xi, yi)
			// Translate camera position.
			op.GeoM.Translate(-g.camX, g.camY)
			// Zoom.
			op.GeoM.Scale(g.camScale, g.camScale)
			// Center.
			op.GeoM.Translate(float64(g.w/2.0), float64(g.h/2.0))

			t.Draw(screen, op)
		}
	}
}
