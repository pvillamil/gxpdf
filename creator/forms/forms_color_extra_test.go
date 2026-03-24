package forms_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coregx/gxpdf/creator/forms"
)

// TestTextField_SetBorderColor covers the border color setter and getter.
func TestTextField_SetBorderColor(t *testing.T) {
	field := forms.NewTextField("f", 0, 0, 100, 20)

	r, g, b := 0.0, 0.0, 0.0
	err := field.SetBorderColor(&r, &g, &b)
	require.NoError(t, err)

	bc := field.BorderColor()
	require.NotNil(t, bc)
	assert.InDelta(t, 0.0, bc[0], 0.001)
	assert.InDelta(t, 0.0, bc[1], 0.001)
	assert.InDelta(t, 0.0, bc[2], 0.001)
}

func TestTextField_SetBorderColor_Nil(t *testing.T) {
	field := forms.NewTextField("f", 0, 0, 100, 20)

	err := field.SetBorderColor(nil, nil, nil)
	require.NoError(t, err)
	assert.Nil(t, field.BorderColor())
}

func TestTextField_SetBorderColor_InvalidRange(t *testing.T) {
	field := forms.NewTextField("f", 0, 0, 100, 20)

	r := 2.0 // out of [0,1]
	g, b := 0.0, 0.0
	err := field.SetBorderColor(&r, &g, &b)
	assert.Error(t, err)
}

func TestTextField_SetFillColor(t *testing.T) {
	field := forms.NewTextField("f", 0, 0, 100, 20)

	r, g, b := 1.0, 1.0, 1.0
	err := field.SetFillColor(&r, &g, &b)
	require.NoError(t, err)

	fc := field.FillColor()
	require.NotNil(t, fc)
	assert.InDelta(t, 1.0, fc[0], 0.001)
	assert.InDelta(t, 1.0, fc[1], 0.001)
	assert.InDelta(t, 1.0, fc[2], 0.001)
}

func TestTextField_SetFillColor_Nil(t *testing.T) {
	field := forms.NewTextField("f", 0, 0, 100, 20)

	err := field.SetFillColor(nil, nil, nil)
	require.NoError(t, err)
	assert.Nil(t, field.FillColor())
}

func TestTextField_SetFillColor_InvalidRange(t *testing.T) {
	field := forms.NewTextField("f", 0, 0, 100, 20)

	r, g, b := 0.0, -0.5, 0.0 // g out of range
	err := field.SetFillColor(&r, &g, &b)
	assert.Error(t, err)
}

// TestDropdown_SetBorderColor covers the dropdown border color methods.
func TestDropdown_SetBorderColor(t *testing.T) {
	dd := forms.NewDropdown("d", 0, 0, 100, 20)
	dd.AddOptions("A", "B")

	r, g, b := 0.5, 0.5, 0.5
	err := dd.SetBorderColor(&r, &g, &b)
	require.NoError(t, err)

	bc := dd.BorderColor()
	require.NotNil(t, bc)
	assert.InDelta(t, 0.5, bc[0], 0.001)
}

func TestDropdown_SetBorderColor_Nil(t *testing.T) {
	dd := forms.NewDropdown("d", 0, 0, 100, 20)
	dd.AddOptions("A")

	err := dd.SetBorderColor(nil, nil, nil)
	require.NoError(t, err)
	assert.Nil(t, dd.BorderColor())
}

func TestDropdown_SetBorderColor_Invalid(t *testing.T) {
	dd := forms.NewDropdown("d", 0, 0, 100, 20)
	dd.AddOptions("A")

	r, g, b := 1.5, 0.0, 0.0 // r out of range
	err := dd.SetBorderColor(&r, &g, &b)
	assert.Error(t, err)
}

func TestDropdown_SetFillColor(t *testing.T) {
	dd := forms.NewDropdown("d", 0, 0, 100, 20)
	dd.AddOptions("A", "B")

	r, g, b := 0.9, 0.9, 0.9
	err := dd.SetFillColor(&r, &g, &b)
	require.NoError(t, err)

	fc := dd.FillColor()
	require.NotNil(t, fc)
	assert.InDelta(t, 0.9, fc[0], 0.001)
}

func TestDropdown_SetFillColor_Nil(t *testing.T) {
	dd := forms.NewDropdown("d", 0, 0, 100, 20)
	dd.AddOptions("A")

	err := dd.SetFillColor(nil, nil, nil)
	require.NoError(t, err)
	assert.Nil(t, dd.FillColor())
}

func TestDropdown_SetFillColor_Invalid(t *testing.T) {
	dd := forms.NewDropdown("d", 0, 0, 100, 20)
	dd.AddOptions("A")

	r, g, b := 0.0, 0.0, -1.0 // b out of range
	err := dd.SetFillColor(&r, &g, &b)
	assert.Error(t, err)
}

// TestListBox_SetBorderColor covers the listbox border color methods.
func TestListBox_SetBorderColor(t *testing.T) {
	lb := forms.NewListBox("lb", 0, 0, 100, 60)
	lb.AddOptions("X", "Y")

	r, g, b := 0.0, 0.0, 1.0
	err := lb.SetBorderColor(&r, &g, &b)
	require.NoError(t, err)

	bc := lb.BorderColor()
	require.NotNil(t, bc)
	assert.InDelta(t, 1.0, bc[2], 0.001)
}

func TestListBox_SetBorderColor_Nil(t *testing.T) {
	lb := forms.NewListBox("lb", 0, 0, 100, 60)
	lb.AddOptions("X")

	err := lb.SetBorderColor(nil, nil, nil)
	require.NoError(t, err)
	assert.Nil(t, lb.BorderColor())
}

func TestListBox_SetBorderColor_Invalid(t *testing.T) {
	lb := forms.NewListBox("lb", 0, 0, 100, 60)
	lb.AddOptions("X")

	r, g, b := 0.0, 2.0, 0.0 // g out of range
	err := lb.SetBorderColor(&r, &g, &b)
	assert.Error(t, err)
}

func TestListBox_SetFillColor(t *testing.T) {
	lb := forms.NewListBox("lb", 0, 0, 100, 60)
	lb.AddOptions("X", "Y")

	r, g, b := 0.8, 0.8, 0.8
	err := lb.SetFillColor(&r, &g, &b)
	require.NoError(t, err)

	fc := lb.FillColor()
	require.NotNil(t, fc)
	assert.InDelta(t, 0.8, fc[0], 0.001)
}

func TestListBox_SetFillColor_Nil(t *testing.T) {
	lb := forms.NewListBox("lb", 0, 0, 100, 60)
	lb.AddOptions("X")

	err := lb.SetFillColor(nil, nil, nil)
	require.NoError(t, err)
	assert.Nil(t, lb.FillColor())
}

func TestListBox_SetFillColor_Invalid(t *testing.T) {
	lb := forms.NewListBox("lb", 0, 0, 100, 60)
	lb.AddOptions("X")

	r, g, b := 0.0, 0.0, 1.5 // b out of range
	err := lb.SetFillColor(&r, &g, &b)
	assert.Error(t, err)
}

// TestRadioGroup_Rect covers the Rect() method.
func TestRadioGroup_Rect(t *testing.T) {
	rg := forms.NewRadioGroup("size")

	// Empty group returns zeroed rect.
	emptyRect := rg.Rect()
	assert.Equal(t, [4]float64{0, 0, 0, 0}, emptyRect)

	// Add an option and verify the rect matches the first option.
	rg.AddOption("sm", 10, 20, "Small")
	rect := rg.Rect()
	assert.Equal(t, 10.0, rect[0], "x should match first option x")
	assert.Equal(t, 20.0, rect[1], "y should match first option y")
}
