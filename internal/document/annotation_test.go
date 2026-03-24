package document

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- LinkAnnotation ----

func TestNewLinkAnnotation(t *testing.T) {
	rect := [4]float64{10, 20, 110, 40}
	a := NewLinkAnnotation(rect, "https://example.com")

	require.NotNil(t, a)
	assert.Equal(t, rect, a.Rect)
	assert.Equal(t, "https://example.com", a.URI)
	assert.False(t, a.IsInternal)
	assert.Equal(t, -1, a.DestPage)
	assert.Equal(t, 0.0, a.BorderWidth)
}

func TestNewInternalLinkAnnotation(t *testing.T) {
	rect := [4]float64{0, 0, 100, 20}
	a := NewInternalLinkAnnotation(rect, 3)

	require.NotNil(t, a)
	assert.Equal(t, rect, a.Rect)
	assert.Equal(t, "", a.URI)
	assert.True(t, a.IsInternal)
	assert.Equal(t, 3, a.DestPage)
}

func TestLinkAnnotation_Validate(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *LinkAnnotation
		wantError bool
		errTarget error
	}{
		{
			name: "valid external link",
			setup: func() *LinkAnnotation {
				return NewLinkAnnotation([4]float64{0, 0, 100, 20}, "https://go.dev")
			},
		},
		{
			name: "valid internal link",
			setup: func() *LinkAnnotation {
				return NewInternalLinkAnnotation([4]float64{0, 0, 100, 20}, 0)
			},
		},
		{
			name: "invalid rect x1>=x2",
			setup: func() *LinkAnnotation {
				a := NewLinkAnnotation([4]float64{100, 0, 100, 20}, "https://go.dev")
				return a
			},
			wantError: true,
			errTarget: ErrInvalidAnnotationRect,
		},
		{
			name: "invalid rect y1>=y2",
			setup: func() *LinkAnnotation {
				return NewLinkAnnotation([4]float64{0, 20, 100, 20}, "https://go.dev")
			},
			wantError: true,
			errTarget: ErrInvalidAnnotationRect,
		},
		{
			name: "external link empty URI",
			setup: func() *LinkAnnotation {
				a := NewLinkAnnotation([4]float64{0, 0, 100, 20}, "")
				return a
			},
			wantError: true,
			errTarget: ErrEmptyURI,
		},
		{
			name: "internal link negative dest page",
			setup: func() *LinkAnnotation {
				return NewInternalLinkAnnotation([4]float64{0, 0, 100, 20}, -1)
			},
			wantError: true,
			errTarget: ErrInvalidDestPage,
		},
		{
			name: "negative border width",
			setup: func() *LinkAnnotation {
				a := NewLinkAnnotation([4]float64{0, 0, 100, 20}, "https://go.dev")
				a.BorderWidth = -1.0
				return a
			},
			wantError: true,
			errTarget: ErrInvalidBorderWidth,
		},
		{
			name: "valid link with custom border width",
			setup: func() *LinkAnnotation {
				a := NewLinkAnnotation([4]float64{0, 0, 100, 20}, "https://go.dev")
				a.BorderWidth = 2.0
				return a
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.setup()
			err := a.Validate()
			if tt.wantError {
				require.Error(t, err)
				if tt.errTarget != nil {
					assert.True(t, errors.Is(err, tt.errTarget),
						"expected error %v, got %v", tt.errTarget, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- TextAnnotation ----

func TestNewTextAnnotation(t *testing.T) {
	rect := [4]float64{10, 10, 30, 30}
	a := NewTextAnnotation(rect, "Note content", "Alice")

	require.NotNil(t, a)
	assert.Equal(t, rect, a.Rect)
	assert.Equal(t, "Note content", a.Contents)
	assert.Equal(t, "Alice", a.Title)
	assert.Equal(t, [3]float64{1, 1, 0}, a.Color)
	assert.False(t, a.Open)
}

func TestTextAnnotation_SetColor(t *testing.T) {
	a := NewTextAnnotation([4]float64{0, 0, 20, 20}, "text", "")
	a.SetColor([3]float64{0.5, 0.5, 0.5})
	assert.Equal(t, [3]float64{0.5, 0.5, 0.5}, a.Color)
}

func TestTextAnnotation_SetOpen(t *testing.T) {
	a := NewTextAnnotation([4]float64{0, 0, 20, 20}, "text", "")
	assert.False(t, a.Open)
	a.SetOpen(true)
	assert.True(t, a.Open)
	a.SetOpen(false)
	assert.False(t, a.Open)
}

func TestTextAnnotation_Validate(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *TextAnnotation
		wantError bool
		errTarget error
	}{
		{
			name: "valid",
			setup: func() *TextAnnotation {
				return NewTextAnnotation([4]float64{0, 0, 20, 20}, "content", "author")
			},
		},
		{
			name: "invalid rect",
			setup: func() *TextAnnotation {
				return NewTextAnnotation([4]float64{20, 0, 10, 20}, "content", "author")
			},
			wantError: true,
			errTarget: ErrInvalidAnnotationRect,
		},
		{
			name: "invalid color component > 1",
			setup: func() *TextAnnotation {
				a := NewTextAnnotation([4]float64{0, 0, 20, 20}, "content", "author")
				a.Color = [3]float64{1.5, 0, 0}
				return a
			},
			wantError: true,
			errTarget: ErrInvalidColor,
		},
		{
			name: "invalid color component < 0",
			setup: func() *TextAnnotation {
				a := NewTextAnnotation([4]float64{0, 0, 20, 20}, "content", "author")
				a.Color = [3]float64{0, -0.1, 0}
				return a
			},
			wantError: true,
			errTarget: ErrInvalidColor,
		},
		{
			name: "empty contents allowed",
			setup: func() *TextAnnotation {
				return NewTextAnnotation([4]float64{0, 0, 20, 20}, "", "")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.setup()
			err := a.Validate()
			if tt.wantError {
				require.Error(t, err)
				if tt.errTarget != nil {
					assert.True(t, errors.Is(err, tt.errTarget))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- MarkupAnnotation ----

func TestNewMarkupAnnotation(t *testing.T) {
	rect := [4]float64{100, 650, 300, 670}
	quads := [][8]float64{{100, 670, 300, 670, 100, 650, 300, 650}}
	a := NewMarkupAnnotation(AnnotationTypeHighlight, rect, quads)

	require.NotNil(t, a)
	assert.Equal(t, AnnotationTypeHighlight, a.Type)
	assert.Equal(t, rect, a.Rect)
	assert.Equal(t, quads, a.QuadPoints)
	assert.Equal(t, [3]float64{1, 1, 0}, a.Color)
	assert.Equal(t, "", a.Title)
	assert.Equal(t, "", a.Contents)
}

func TestMarkupAnnotation_SetColor(t *testing.T) {
	a := NewMarkupAnnotation(AnnotationTypeUnderline, [4]float64{0, 0, 100, 20},
		[][8]float64{{0, 20, 100, 20, 0, 0, 100, 0}})
	a.SetColor([3]float64{0, 0, 1})
	assert.Equal(t, [3]float64{0, 0, 1}, a.Color)
}

func TestMarkupAnnotation_SetAuthor(t *testing.T) {
	a := NewMarkupAnnotation(AnnotationTypeStrikeOut, [4]float64{0, 0, 100, 20},
		[][8]float64{{0, 20, 100, 20, 0, 0, 100, 0}})
	a.SetAuthor("Bob")
	assert.Equal(t, "Bob", a.Title)
}

func TestMarkupAnnotation_SetContents(t *testing.T) {
	a := NewMarkupAnnotation(AnnotationTypeHighlight, [4]float64{0, 0, 100, 20},
		[][8]float64{{0, 20, 100, 20, 0, 0, 100, 0}})
	a.SetContents("important")
	assert.Equal(t, "important", a.Contents)
}

func TestMarkupAnnotation_Validate(t *testing.T) {
	validRect := [4]float64{0, 0, 200, 20}
	validQuads := [][8]float64{{0, 20, 200, 20, 0, 0, 200, 0}}

	tests := []struct {
		name      string
		setup     func() *MarkupAnnotation
		wantError bool
		errTarget error
	}{
		{
			name: "valid highlight",
			setup: func() *MarkupAnnotation {
				return NewMarkupAnnotation(AnnotationTypeHighlight, validRect, validQuads)
			},
		},
		{
			name: "valid underline",
			setup: func() *MarkupAnnotation {
				return NewMarkupAnnotation(AnnotationTypeUnderline, validRect, validQuads)
			},
		},
		{
			name: "valid strikeout",
			setup: func() *MarkupAnnotation {
				return NewMarkupAnnotation(AnnotationTypeStrikeOut, validRect, validQuads)
			},
		},
		{
			name: "invalid rect",
			setup: func() *MarkupAnnotation {
				return NewMarkupAnnotation(AnnotationTypeHighlight, [4]float64{200, 0, 100, 20}, validQuads)
			},
			wantError: true,
			errTarget: ErrInvalidAnnotationRect,
		},
		{
			name: "invalid color",
			setup: func() *MarkupAnnotation {
				a := NewMarkupAnnotation(AnnotationTypeHighlight, validRect, validQuads)
				a.Color = [3]float64{2, 0, 0}
				return a
			},
			wantError: true,
			errTarget: ErrInvalidColor,
		},
		{
			name: "missing quad points",
			setup: func() *MarkupAnnotation {
				return NewMarkupAnnotation(AnnotationTypeHighlight, validRect, [][8]float64{})
			},
			wantError: true,
			errTarget: ErrMissingQuadPoints,
		},
		{
			name: "nil quad points",
			setup: func() *MarkupAnnotation {
				return NewMarkupAnnotation(AnnotationTypeHighlight, validRect, nil)
			},
			wantError: true,
			errTarget: ErrMissingQuadPoints,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.setup()
			err := a.Validate()
			if tt.wantError {
				require.Error(t, err)
				if tt.errTarget != nil {
					assert.True(t, errors.Is(err, tt.errTarget))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ---- StampAnnotation ----

func TestNewStampAnnotation(t *testing.T) {
	rect := [4]float64{300, 700, 400, 750}
	a := NewStampAnnotation(rect, StampApproved)

	require.NotNil(t, a)
	assert.Equal(t, rect, a.Rect)
	assert.Equal(t, "Approved", a.Name)
	assert.Equal(t, [3]float64{1, 0, 0}, a.Color)
	assert.Equal(t, "", a.Title)
	assert.Equal(t, "", a.Contents)
}

func TestStampAnnotation_SetColor(t *testing.T) {
	a := NewStampAnnotation([4]float64{0, 0, 100, 50}, StampDraft)
	a.SetColor([3]float64{0, 0.5, 0})
	assert.Equal(t, [3]float64{0, 0.5, 0}, a.Color)
}

func TestStampAnnotation_SetAuthorContents(t *testing.T) {
	a := NewStampAnnotation([4]float64{0, 0, 100, 50}, StampConfidential)
	a.SetAuthor("Carol")
	a.SetContents("Do not distribute")
	assert.Equal(t, "Carol", a.Title)
	assert.Equal(t, "Do not distribute", a.Contents)
}

func TestStampAnnotation_Validate(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *StampAnnotation
		wantError bool
		errTarget error
	}{
		{
			name: "valid approved stamp",
			setup: func() *StampAnnotation {
				return NewStampAnnotation([4]float64{0, 0, 100, 50}, StampApproved)
			},
		},
		{
			name: "valid draft stamp",
			setup: func() *StampAnnotation {
				return NewStampAnnotation([4]float64{0, 0, 100, 50}, StampDraft)
			},
		},
		{
			name: "invalid rect",
			setup: func() *StampAnnotation {
				return NewStampAnnotation([4]float64{100, 0, 50, 50}, StampApproved)
			},
			wantError: true,
			errTarget: ErrInvalidAnnotationRect,
		},
		{
			name: "invalid color",
			setup: func() *StampAnnotation {
				a := NewStampAnnotation([4]float64{0, 0, 100, 50}, StampApproved)
				a.Color = [3]float64{-0.1, 0, 0}
				return a
			},
			wantError: true,
			errTarget: ErrInvalidColor,
		},
		{
			name: "missing name",
			setup: func() *StampAnnotation {
				a := NewStampAnnotation([4]float64{0, 0, 100, 50}, StampApproved)
				a.Name = ""
				return a
			},
			wantError: true,
			errTarget: ErrMissingStampName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.setup()
			err := a.Validate()
			if tt.wantError {
				require.Error(t, err)
				if tt.errTarget != nil {
					assert.True(t, errors.Is(err, tt.errTarget))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStampNames_AllConstants(t *testing.T) {
	// Ensure all stamp name constants are non-empty.
	stamps := []StampName{
		StampApproved, StampNotApproved, StampDraft, StampFinal,
		StampConfidential, StampForComment, StampForPublicRelease,
		StampAsIs, StampDepartmental, StampExperimental,
		StampExpired, StampNotForPublicRelease,
	}
	for _, s := range stamps {
		assert.NotEmpty(t, string(s))
	}
}

func TestAnnotationType_Values(t *testing.T) {
	// Ensure annotation type constants are distinct iota values.
	types := []AnnotationType{
		AnnotationTypeLink,
		AnnotationTypeText,
		AnnotationTypeHighlight,
		AnnotationTypeUnderline,
		AnnotationTypeStrikeOut,
		AnnotationTypeStamp,
	}
	seen := make(map[AnnotationType]bool)
	for _, at := range types {
		assert.False(t, seen[at], "duplicate AnnotationType value: %d", at)
		seen[at] = true
	}
}

func TestIsValidColor(t *testing.T) {
	tests := []struct {
		color [3]float64
		valid bool
	}{
		{[3]float64{0, 0, 0}, true},
		{[3]float64{1, 1, 1}, true},
		{[3]float64{0.5, 0.5, 0.5}, true},
		{[3]float64{-0.01, 0, 0}, false},
		{[3]float64{0, 1.01, 0}, false},
		{[3]float64{0, 0, 2.0}, false},
	}

	for _, tt := range tests {
		result := isValidColor(tt.color)
		assert.Equal(t, tt.valid, result, "color %v", tt.color)
	}
}
