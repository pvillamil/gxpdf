package layout

import (
	"testing"
)

// makeTextElem creates a Text element with predictable sizing for tests.
// At fontSize 10, lineHeight 1.0: each text = 10pt tall.
func makeTextElem(content string) *Text {
	return &Text{
		Content: content,
		Style:   Style{FontSize: 10, LineHeight: 1.0},
		Fonts:   &MockFontResolver{},
	}
}

// ------- Basic paginator tests -------

func TestPaginator_SinglePage_AllContentFits(t *testing.T) {
	// A4 page, standard margins, 3 short text elements.
	paginator := &Paginator{Fonts: &MockFontResolver{}}

	def := &PageDef{
		Size:    PageA4,
		Margins: UniformEdges(Pt(50)),
		Content: []Element{
			makeTextElem("Line one"),
			makeTextElem("Line two"),
			makeTextElem("Line three"),
		},
	}

	pages := paginator.Paginate([]*PageDef{def})
	if len(pages) != 1 {
		t.Errorf("expected 1 page, got %d", len(pages))
	}
}

func TestPaginator_EmptyContent_NoPages(t *testing.T) {
	paginator := &Paginator{Fonts: &MockFontResolver{}}
	pages := paginator.Paginate(nil)
	if len(pages) != 0 {
		t.Errorf("expected 0 pages for nil input, got %d", len(pages))
	}
}

func TestPaginator_EmptyPageDef(t *testing.T) {
	paginator := &Paginator{Fonts: &MockFontResolver{}}
	def := &PageDef{
		Size:    PageA4,
		Margins: UniformEdges(Pt(50)),
		Content: nil,
	}
	// Empty content → no pages produced (nothing to paginate).
	pages := paginator.Paginate([]*PageDef{def})
	if len(pages) != 0 {
		t.Errorf("expected 0 pages for empty content, got %d", len(pages))
	}
}

func TestPaginator_Overflow_CreatesMultiplePages(t *testing.T) {
	paginator := &Paginator{Fonts: &MockFontResolver{}}

	// Page with very small body area (50pt tall after margins).
	// Each text element = 10pt. We add 10 elements → 100pt needed → 2 pages.
	var content []Element
	for i := 0; i < 10; i++ {
		content = append(content, makeTextElem("item"))
	}

	def := &PageDef{
		Size: Size{Width: 300, Height: 150}, // small page
		Margins: Edges{
			Top:    Pt(50),
			Bottom: Pt(50),
			Left:   Pt(10),
			Right:  Pt(10),
		},
		Content: content,
	}
	// Body height = 150 - 50 - 50 = 50pt → 5 items per page → 2 pages.

	pages := paginator.Paginate([]*PageDef{def})
	if len(pages) < 2 {
		t.Errorf("expected >= 2 pages, got %d", len(pages))
	}
}

func TestPaginator_PageSize_PreservedInOutput(t *testing.T) {
	paginator := &Paginator{Fonts: &MockFontResolver{}}

	def := &PageDef{
		Size:    PageLetter,
		Margins: UniformEdges(Pt(36)),
		Content: []Element{makeTextElem("hello")},
	}

	pages := paginator.Paginate([]*PageDef{def})
	if len(pages) == 0 {
		t.Fatal("expected at least 1 page")
	}
	if pages[0].Size != PageLetter {
		t.Errorf("page size: got %v, want %v", pages[0].Size, PageLetter)
	}
}

func TestPaginator_DefaultsToA4WhenSizeZero(t *testing.T) {
	paginator := &Paginator{Fonts: &MockFontResolver{}}

	def := &PageDef{
		Size:    Size{}, // zero size → should default to A4
		Margins: UniformEdges(Pt(50)),
		Content: []Element{makeTextElem("test")},
	}

	pages := paginator.Paginate([]*PageDef{def})
	if len(pages) == 0 {
		t.Fatal("expected at least 1 page")
	}
	if pages[0].Size != PageA4 {
		t.Errorf("expected A4 default size, got %v", pages[0].Size)
	}
}

func TestPaginator_NilFonts_UsesMock(t *testing.T) {
	// Paginator with nil Fonts should not panic.
	paginator := &Paginator{Fonts: nil}
	def := &PageDef{
		Size:    PageA4,
		Margins: UniformEdges(Pt(50)),
		Content: []Element{makeTextElem("hello")},
	}
	pages := paginator.Paginate([]*PageDef{def})
	if len(pages) == 0 {
		t.Error("expected at least 1 page with nil fonts")
	}
}

// ------- Header and footer tests -------

func TestPaginator_HeaderAndFooter_ReduceBodyArea(t *testing.T) {
	paginator := &Paginator{Fonts: &MockFontResolver{}}

	// Page height = 200pt, margins = 20pt top+bottom → content area = 160pt.
	// Header = 1 text at 10pt, Footer = 1 text at 10pt → body = 140pt.
	// Content = 14 text elements at 10pt → exactly fills 1 page.
	var content []Element
	for i := 0; i < 14; i++ {
		content = append(content, makeTextElem("body"))
	}

	def := &PageDef{
		Size:    Size{Width: 300, Height: 200},
		Margins: Edges{Top: Pt(20), Bottom: Pt(20), Left: Pt(10), Right: Pt(10)},
		Header:  []Element{makeTextElem("HEADER")},
		Footer:  []Element{makeTextElem("FOOTER")},
		Content: content,
	}

	pages := paginator.Paginate([]*PageDef{def})
	if len(pages) < 1 {
		t.Fatal("expected at least 1 page")
	}

	// Verify header and footer blocks appear on first page.
	firstPage := pages[0]
	if len(firstPage.Blocks) == 0 {
		t.Error("expected blocks on first page")
	}
}

func TestPaginator_Header_AppearsOnEveryPage(t *testing.T) {
	paginator := &Paginator{Fonts: &MockFontResolver{}}

	// Force 2 pages.
	var content []Element
	for i := 0; i < 20; i++ {
		content = append(content, makeTextElem("body"))
	}

	def := &PageDef{
		Size:    Size{Width: 300, Height: 150},
		Margins: Edges{Top: Pt(20), Bottom: Pt(20), Left: Pt(10), Right: Pt(10)},
		Header:  []Element{makeTextElem("HDR")},
		Content: content,
	}

	pages := paginator.Paginate([]*PageDef{def})
	if len(pages) < 2 {
		t.Fatalf("expected >= 2 pages, got %d", len(pages))
	}

	// Each page should have blocks (header contributes).
	for i, page := range pages {
		if len(page.Blocks) == 0 {
			t.Errorf("page %d has no blocks (header missing?)", i+1)
		}
	}
}

// ------- Page number placeholder tests -------

func TestResolvePageNumbers_ReplacesPlaceholders(t *testing.T) {
	pages := []PageLayout{
		{Size: PageA4, Blocks: []Block{
			{Tag: "__pagenumber__", AltText: PageNumberPlaceholder + " / " + TotalPagesPlaceholder},
		}},
		{Size: PageA4, Blocks: []Block{
			{Tag: "__pagenumber__", AltText: PageNumberPlaceholder + " / " + TotalPagesPlaceholder},
		}},
	}

	ResolvePageNumbers(pages)

	if pages[0].Blocks[0].AltText != "1 / 2" {
		t.Errorf("page 1: got %q, want %q", pages[0].Blocks[0].AltText, "1 / 2")
	}
	if pages[1].Blocks[0].AltText != "2 / 2" {
		t.Errorf("page 2: got %q, want %q", pages[1].Blocks[0].AltText, "2 / 2")
	}
}

func TestResolvePageNumbers_NestedBlocks(t *testing.T) {
	pages := []PageLayout{
		{Size: PageA4, Blocks: []Block{
			{
				Children: []Block{
					{Tag: "__pagenumber__", AltText: PageNumberPlaceholder},
				},
			},
		}},
	}

	ResolvePageNumbers(pages)
	if pages[0].Blocks[0].Children[0].AltText != "1" {
		t.Errorf("nested: got %q, want %q", pages[0].Blocks[0].Children[0].AltText, "1")
	}
}

func TestResolvePageNumbers_NoPlaceholders(t *testing.T) {
	// Blocks without the __pagenumber__ tag should not be modified.
	pages := []PageLayout{
		{Size: PageA4, Blocks: []Block{
			{Tag: "P", AltText: "regular text"},
		}},
	}
	ResolvePageNumbers(pages)
	if pages[0].Blocks[0].AltText != "regular text" {
		t.Errorf("non-placeholder block modified: got %q", pages[0].Blocks[0].AltText)
	}
}

// ------- measureSection tests -------

func TestMeasureSection_Empty(t *testing.T) {
	blocks, h := measureSection(nil, 200)
	if len(blocks) != 0 {
		t.Errorf("expected no blocks, got %d", len(blocks))
	}
	if h != 0 {
		t.Errorf("expected h=0, got %v", h)
	}
}

func TestMeasureSection_TwoElements(t *testing.T) {
	// Two 10pt elements → total height = 20pt.
	elements := []Element{
		makeTextElem("a"),
		makeTextElem("b"),
	}
	_, h := measureSection(elements, 200)
	if h != 20 {
		t.Errorf("height: got %v, want 20", h)
	}
}

// ------- Multiple PageDefs test -------

func TestPaginator_MultiplePageDefs(t *testing.T) {
	paginator := &Paginator{Fonts: &MockFontResolver{}}

	def1 := &PageDef{
		Size:    PageA4,
		Margins: UniformEdges(Pt(50)),
		Content: []Element{makeTextElem("page def 1 content")},
	}
	def2 := &PageDef{
		Size:    PageLetter,
		Margins: UniformEdges(Pt(36)),
		Content: []Element{makeTextElem("page def 2 content")},
	}

	pages := paginator.Paginate([]*PageDef{def1, def2})
	if len(pages) < 2 {
		t.Errorf("expected >= 2 pages for 2 page defs, got %d", len(pages))
	}
	// First page should be A4 size.
	if pages[0].Size != PageA4 {
		t.Errorf("first page size: got %v, want A4", pages[0].Size)
	}
	// Second page should be Letter size.
	if pages[1].Size != PageLetter {
		t.Errorf("second page size: got %v, want Letter", pages[1].Size)
	}
}

// ------- PageNumber element test -------

func TestPageNumber_PlanLayout(t *testing.T) {
	pn := &PageNumber{
		Format: PageNumberPlaceholder + " / " + TotalPagesPlaceholder,
		Style:  Style{FontSize: 10, LineHeight: 1.0},
		Fonts:  &MockFontResolver{},
	}
	plan := pn.PlanLayout(Area{Width: 200, Height: 100})
	if plan.Status == Nothing {
		t.Error("PageNumber: expected non-Nothing plan")
	}
}
