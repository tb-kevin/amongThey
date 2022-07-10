package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Level represents a Game level.
type Level struct {
	w, h int

	tiles    [][]*Tile // (Y,X) array of tiles
	tileSize int
}

// Tile returns the tile at the provided coordinates, or nil.
func (l *Level) Tile(x, y int) *Tile {
	if x >= 0 && y >= 0 && x < l.w && y < l.h {
		return l.tiles[y][x]
	}
	return nil
}

// Size returns the size of the Level.
func (l *Level) Size() (width, height int) {
	return l.w, l.h
}

// NewLevel returns a new randomly generated Level.
func NewLevel() (*Level, error) {
	l := &Level{
		w:        30,
		h:        30,
		tileSize: 64,
	}

	// Load embedded SpriteSheet.
	ss, err := LoadSpriteSheet(l.tileSize)
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded spritesheet: %s", err)
	}

	// Generate a unique permutation each time.
	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	// Fill each tile with one or more sprites randomly.
	l.tiles = make([][]*Tile, l.h)
	for y := 0; y < l.h; y++ {
		l.tiles[y] = make([]*Tile, l.w)
		for x := 0; x < l.w; x++ {
			t := &Tile{}
			isBorderSpace := x == 0 || y == 0 || x == l.w-1 || y == l.h-1
			val := r.Intn(1000)
			switch {
			case isBorderSpace || val < 275:
				t.AddSprite(ss.Wall)
			case val < 285:
				t.AddSprite(ss.RunnerImage)
			case val < 288:
				t.AddSprite(ss.Crown)
			case val < 289:
				t.AddSprite(ss.Floor)
				t.AddSprite(ss.Tube)
			case val < 290:
				t.AddSprite(ss.Portal)
			default:
				t.AddSprite(ss.Floor)
			}
			l.tiles[y][x] = t
		}
	}

	return l, nil
}
