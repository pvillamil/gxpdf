package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRGB covers the layout.RGB constructor.
func TestRGB(t *testing.T) {
	c := RGB(0.5, 0.25, 0.75)
	assert.InDelta(t, 0.5, c.R, 0.001)
	assert.InDelta(t, 0.25, c.G, 0.001)
	assert.InDelta(t, 0.75, c.B, 0.001)
}

func TestRGB_Black(t *testing.T) {
	c := RGB(0, 0, 0)
	assert.Equal(t, Black, c)
}

func TestRGB_White(t *testing.T) {
	c := RGB(1, 1, 1)
	assert.Equal(t, White, c)
}

// TestRecomputeCursorY covers the internal recomputeCursorY function.
func TestRecomputeCursorY_Empty(t *testing.T) {
	result := recomputeCursorY(nil)
	assert.Equal(t, 0.0, result)
}

func TestRecomputeCursorY_SingleBlock(t *testing.T) {
	blocks := []Block{{Y: 10, Height: 20}}
	result := recomputeCursorY(blocks)
	assert.InDelta(t, 30.0, result, 0.001)
}

func TestRecomputeCursorY_MultipleBlocks(t *testing.T) {
	blocks := []Block{
		{Y: 0, Height: 20},
		{Y: 20, Height: 15},
		{Y: 35, Height: 10},
	}
	result := recomputeCursorY(blocks)
	assert.InDelta(t, 45.0, result, 0.001) // last block: Y=35 + H=10
}
