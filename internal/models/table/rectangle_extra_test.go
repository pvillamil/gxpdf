package table

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRectangle_Right(t *testing.T) {
	r := NewRectangle(10, 20, 30, 40)
	assert.InDelta(t, 40.0, r.Right(), 0.001)
}

func TestRectangle_Top(t *testing.T) {
	r := NewRectangle(10, 20, 30, 40)
	assert.InDelta(t, 60.0, r.Top(), 0.001)
}

func TestRectangle_Bottom(t *testing.T) {
	r := NewRectangle(10, 20, 30, 40)
	assert.InDelta(t, 20.0, r.Bottom(), 0.001)
}

func TestRectangle_Left(t *testing.T) {
	r := NewRectangle(10, 20, 30, 40)
	assert.InDelta(t, 10.0, r.Left(), 0.001)
}

func TestRectangle_Contains(t *testing.T) {
	r := NewRectangle(0, 0, 100, 50)

	assert.True(t, r.Contains(50, 25), "center point should be inside")
	assert.True(t, r.Contains(0, 0), "bottom-left corner should be inside")
	assert.True(t, r.Contains(100, 50), "top-right corner should be inside")
	assert.False(t, r.Contains(-1, 25), "point left of rect should be outside")
	assert.False(t, r.Contains(50, 51), "point above rect should be outside")
	assert.False(t, r.Contains(101, 25), "point right of rect should be outside")
	assert.False(t, r.Contains(50, -1), "point below rect should be outside")
}

func TestRectangle_String(t *testing.T) {
	r := NewRectangle(1, 2, 3, 4)
	s := r.String()
	assert.Contains(t, s, "Rectangle")
	assert.Contains(t, s, "1.00")
	assert.Contains(t, s, "2.00")
}

func TestRectangle_ZeroSize(t *testing.T) {
	r := NewRectangle(5, 5, 0, 0)
	assert.InDelta(t, 5.0, r.Right(), 0.001)
	assert.InDelta(t, 5.0, r.Top(), 0.001)
	assert.InDelta(t, 5.0, r.Left(), 0.001)
	assert.InDelta(t, 5.0, r.Bottom(), 0.001)
}
