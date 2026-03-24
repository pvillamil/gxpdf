package document

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- Page annotation methods ----

func TestPage_AddLinkAnnotation(t *testing.T) {
	page := NewPage(0, A4)

	link := NewLinkAnnotation([4]float64{0, 0, 100, 20}, "https://go.dev")
	err := page.AddLinkAnnotation(link)
	require.NoError(t, err)
	assert.Len(t, page.LinkAnnotations(), 1)

	// nil annotation
	err = page.AddLinkAnnotation(nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNilAnnotation)
}

func TestPage_AddAnnotation_Deprecated(t *testing.T) {
	// AddAnnotation is the deprecated alias for AddLinkAnnotation.
	page := NewPage(0, A4)
	link := NewLinkAnnotation([4]float64{0, 0, 100, 20}, "https://go.dev")
	err := page.AddAnnotation(link)
	require.NoError(t, err)
	assert.Len(t, page.Annotations(), 1)
	assert.Len(t, page.LinkAnnotations(), 1)
}

func TestPage_AddLinkAnnotation_Invalid(t *testing.T) {
	page := NewPage(0, A4)

	// Invalid annotation (bad rect).
	invalid := NewLinkAnnotation([4]float64{100, 0, 50, 20}, "https://go.dev")
	err := page.AddLinkAnnotation(invalid)
	require.Error(t, err)
	assert.Len(t, page.LinkAnnotations(), 0, "invalid annotation should not be added")
}

func TestPage_AddTextAnnotation(t *testing.T) {
	page := NewPage(0, A4)

	note := NewTextAnnotation([4]float64{10, 10, 30, 30}, "Important!", "Alice")
	err := page.AddTextAnnotation(note)
	require.NoError(t, err)
	assert.Len(t, page.TextAnnotations(), 1)

	// nil
	err = page.AddTextAnnotation(nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNilAnnotation)
}

func TestPage_AddTextAnnotation_Invalid(t *testing.T) {
	page := NewPage(0, A4)

	// Bad rect
	invalid := NewTextAnnotation([4]float64{30, 10, 10, 30}, "txt", "")
	err := page.AddTextAnnotation(invalid)
	require.Error(t, err)
	assert.Len(t, page.TextAnnotations(), 0)
}

func TestPage_AddMarkupAnnotation(t *testing.T) {
	page := NewPage(0, A4)

	quads := [][8]float64{{0, 20, 200, 20, 0, 0, 200, 0}}
	markup := NewMarkupAnnotation(AnnotationTypeHighlight, [4]float64{0, 0, 200, 20}, quads)
	err := page.AddMarkupAnnotation(markup)
	require.NoError(t, err)
	assert.Len(t, page.MarkupAnnotations(), 1)

	// nil
	err = page.AddMarkupAnnotation(nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNilAnnotation)
}

func TestPage_AddMarkupAnnotation_Invalid(t *testing.T) {
	page := NewPage(0, A4)

	// No quad points.
	invalid := NewMarkupAnnotation(AnnotationTypeHighlight, [4]float64{0, 0, 200, 20}, nil)
	err := page.AddMarkupAnnotation(invalid)
	require.Error(t, err)
	assert.Len(t, page.MarkupAnnotations(), 0)
}

func TestPage_AddStampAnnotation(t *testing.T) {
	page := NewPage(0, A4)

	stamp := NewStampAnnotation([4]float64{300, 700, 400, 750}, StampApproved)
	err := page.AddStampAnnotation(stamp)
	require.NoError(t, err)
	assert.Len(t, page.StampAnnotations(), 1)

	// nil
	err = page.AddStampAnnotation(nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNilAnnotation)
}

func TestPage_AddStampAnnotation_Invalid(t *testing.T) {
	page := NewPage(0, A4)

	// Empty stamp name.
	invalid := NewStampAnnotation([4]float64{0, 0, 100, 50}, StampApproved)
	invalid.Name = ""
	err := page.AddStampAnnotation(invalid)
	require.Error(t, err)
	assert.Len(t, page.StampAnnotations(), 0)
}

func TestPage_AddFormField(t *testing.T) {
	page := NewPage(0, A4)

	field := NewFormField("Tx", "name", [4]float64{100, 700, 300, 720})
	err := page.AddFormField(field)
	require.NoError(t, err)
	assert.Len(t, page.FormFields(), 1)

	// nil
	err = page.AddFormField(nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrNilFormField)
}

func TestPage_AddFormField_Invalid(t *testing.T) {
	page := NewPage(0, A4)

	// Empty field name.
	invalid := NewFormField("Tx", "", [4]float64{0, 0, 100, 20})
	err := page.AddFormField(invalid)
	require.Error(t, err)
	assert.Len(t, page.FormFields(), 0)
}

func TestPage_AnnotationCount(t *testing.T) {
	page := NewPage(0, A4)
	assert.Equal(t, 0, page.AnnotationCount())

	link := NewLinkAnnotation([4]float64{0, 0, 100, 20}, "https://go.dev")
	page.AddLinkAnnotation(link)
	assert.Equal(t, 1, page.AnnotationCount())

	note := NewTextAnnotation([4]float64{10, 10, 30, 30}, "note", "author")
	page.AddTextAnnotation(note)
	assert.Equal(t, 2, page.AnnotationCount())

	quads := [][8]float64{{0, 20, 100, 20, 0, 0, 100, 0}}
	markup := NewMarkupAnnotation(AnnotationTypeHighlight, [4]float64{0, 0, 100, 20}, quads)
	page.AddMarkupAnnotation(markup)
	assert.Equal(t, 3, page.AnnotationCount())

	stamp := NewStampAnnotation([4]float64{0, 0, 100, 50}, StampDraft)
	page.AddStampAnnotation(stamp)
	assert.Equal(t, 4, page.AnnotationCount())

	field := NewFormField("Tx", "f1", [4]float64{0, 0, 100, 20})
	page.AddFormField(field)
	assert.Equal(t, 5, page.AnnotationCount())
}

func TestPage_ClearAnnotations(t *testing.T) {
	page := NewPage(0, A4)

	// Populate all annotation types.
	page.AddLinkAnnotation(NewLinkAnnotation([4]float64{0, 0, 100, 20}, "https://go.dev"))
	page.AddTextAnnotation(NewTextAnnotation([4]float64{10, 10, 30, 30}, "note", ""))
	quads := [][8]float64{{0, 20, 100, 20, 0, 0, 100, 0}}
	page.AddMarkupAnnotation(NewMarkupAnnotation(AnnotationTypeHighlight, [4]float64{0, 0, 100, 20}, quads))
	page.AddStampAnnotation(NewStampAnnotation([4]float64{0, 0, 100, 50}, StampDraft))
	page.AddFormField(NewFormField("Btn", "btn1", [4]float64{0, 0, 100, 20}))

	assert.Equal(t, 5, page.AnnotationCount())

	page.ClearAnnotations()
	assert.Equal(t, 0, page.AnnotationCount())

	// Verify each collection is empty.
	assert.Empty(t, page.LinkAnnotations())
	assert.Empty(t, page.TextAnnotations())
	assert.Empty(t, page.MarkupAnnotations())
	assert.Empty(t, page.StampAnnotations())
	assert.Empty(t, page.FormFields())

	// Can add again after clearing.
	page.AddLinkAnnotation(NewLinkAnnotation([4]float64{0, 0, 100, 20}, "https://go.dev"))
	assert.Equal(t, 1, page.AnnotationCount())
}

func TestPage_AnnotationCollections_AreCopies(t *testing.T) {
	page := NewPage(0, A4)
	page.AddLinkAnnotation(NewLinkAnnotation([4]float64{0, 0, 100, 20}, "https://go.dev"))
	page.AddTextAnnotation(NewTextAnnotation([4]float64{10, 10, 30, 30}, "n", ""))
	quads := [][8]float64{{0, 20, 100, 20, 0, 0, 100, 0}}
	page.AddMarkupAnnotation(NewMarkupAnnotation(AnnotationTypeHighlight, [4]float64{0, 0, 100, 20}, quads))
	page.AddStampAnnotation(NewStampAnnotation([4]float64{0, 0, 100, 50}, StampApproved))
	page.AddFormField(NewFormField("Ch", "choice", [4]float64{0, 0, 100, 20}))

	// Modify returned slices.
	links := page.LinkAnnotations()
	links[0] = nil
	assert.NotNil(t, page.LinkAnnotations()[0])

	texts := page.TextAnnotations()
	texts[0] = nil
	assert.NotNil(t, page.TextAnnotations()[0])

	markups := page.MarkupAnnotations()
	markups[0] = nil
	assert.NotNil(t, page.MarkupAnnotations()[0])

	stamps := page.StampAnnotations()
	stamps[0] = nil
	assert.NotNil(t, page.StampAnnotations()[0])

	fields := page.FormFields()
	fields[0] = nil
	assert.NotNil(t, page.FormFields()[0])
}

func TestPage_Validate_WithAnnotations(t *testing.T) {
	// Valid page with all annotation types passes validation.
	page := NewPage(0, A4)

	link := NewLinkAnnotation([4]float64{0, 0, 100, 20}, "https://go.dev")
	page.AddLinkAnnotation(link)

	note := NewTextAnnotation([4]float64{10, 10, 30, 30}, "note", "author")
	page.AddTextAnnotation(note)

	quads := [][8]float64{{0, 20, 100, 20, 0, 0, 100, 0}}
	markup := NewMarkupAnnotation(AnnotationTypeHighlight, [4]float64{0, 0, 100, 20}, quads)
	page.AddMarkupAnnotation(markup)

	stamp := NewStampAnnotation([4]float64{0, 0, 100, 50}, StampFinal)
	page.AddStampAnnotation(stamp)

	field := NewFormField("Sig", "sig1", [4]float64{0, 0, 150, 30})
	page.AddFormField(field)

	err := page.Validate()
	assert.NoError(t, err)
}

func TestPage_Validate_WithNilAnnotations(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*Page)
	}{
		{
			"nil link annotation",
			func(p *Page) { p.linkAnnotations = append(p.linkAnnotations, nil) },
		},
		{
			"nil text annotation",
			func(p *Page) { p.textAnnotations = append(p.textAnnotations, nil) },
		},
		{
			"nil markup annotation",
			func(p *Page) { p.markupAnnotations = append(p.markupAnnotations, nil) },
		},
		{
			"nil stamp annotation",
			func(p *Page) { p.stampAnnotations = append(p.stampAnnotations, nil) },
		},
		{
			"nil form field",
			func(p *Page) { p.formFields = append(p.formFields, nil) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := NewPage(0, A4)
			tt.setup(page)
			err := page.Validate()
			assert.Error(t, err, "nil annotation should cause validation failure")
		})
	}
}

func TestPage_MultipleAnnotationsOfSameType(t *testing.T) {
	page := NewPage(0, A4)

	for i := 0; i < 5; i++ {
		link := NewLinkAnnotation([4]float64{float64(i * 10), 0, float64(i*10 + 50), 20}, "https://go.dev")
		require.NoError(t, page.AddLinkAnnotation(link))
	}

	assert.Equal(t, 5, page.AnnotationCount())
	assert.Len(t, page.LinkAnnotations(), 5)
}
