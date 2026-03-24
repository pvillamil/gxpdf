package layout

import (
	"strings"
	"testing"
)

// mockFont is a standard FontRef used across text tests.
var mockFont = DefaultFont()

// ------- MockFontResolver tests -------

func TestMockFontResolver_MeasureString(t *testing.T) {
	m := &MockFontResolver{}
	// "Hello" = 5 runes, size 12 → 5 * 12 * 0.5 = 30
	got := m.MeasureString(mockFont, "Hello", 12)
	want := float64(5) * 12 * 0.5
	if got != want {
		t.Errorf("MeasureString = %v, want %v", got, want)
	}
}

func TestMockFontResolver_LineHeight(t *testing.T) {
	m := &MockFontResolver{}
	if got := m.LineHeight(mockFont, 10); got != 12 {
		t.Errorf("LineHeight = %v, want 12", got)
	}
}

func TestMockFontResolver_Ascender(t *testing.T) {
	m := &MockFontResolver{}
	if got := m.Ascender(mockFont, 10); got != 8 {
		t.Errorf("Ascender = %v, want 8", got)
	}
}

func TestMockFontResolver_Descender(t *testing.T) {
	m := &MockFontResolver{}
	if got := m.Descender(mockFont, 10); got != 2 {
		t.Errorf("Descender = %v, want 2", got)
	}
}

func TestMockFontResolver_LineBreak_ShortText(t *testing.T) {
	m := &MockFontResolver{}
	// "Hi" at size 10 → width = 2*10*0.5 = 10. maxWidth 100 → single line.
	lines := m.LineBreak(mockFont, "Hi", 10, 100)
	if len(lines) != 1 {
		t.Errorf("expected 1 line, got %d: %v", len(lines), lines)
	}
	if lines[0] != "Hi" {
		t.Errorf("expected 'Hi', got %q", lines[0])
	}
}

func TestMockFontResolver_LineBreak_Wraps(t *testing.T) {
	m := &MockFontResolver{}
	// "one two three" at size 10, each word 3-5 chars * 0.5 * 10 = 15-25pt.
	// maxWidth = 30pt. "one" = 15pt, "one two" = 35pt > 30 → break.
	lines := m.LineBreak(mockFont, "one two three", 10, 30)
	if len(lines) < 2 {
		t.Errorf("expected wrapping (>1 line), got %d: %v", len(lines), lines)
	}
}

func TestMockFontResolver_LineBreak_EmptyText(t *testing.T) {
	m := &MockFontResolver{}
	lines := m.LineBreak(mockFont, "", 12, 100)
	if len(lines) != 1 || lines[0] != "" {
		t.Errorf("empty text: got %v, want [\"\"]", lines)
	}
}

// ------- Text.PlanLayout tests -------

func TestText_PlanLayout_SingleLine(t *testing.T) {
	// "Hi" at fontSize 10, lineHeight 1.0 → consumes 10pt.
	txt := &Text{
		Content: "Hi",
		Style:   Style{FontSize: 10, LineHeight: 1.0},
		Fonts:   &MockFontResolver{},
	}
	plan := txt.PlanLayout(Area{Width: 200, Height: 100})
	if plan.Status != Full {
		t.Errorf("status: got %v, want Full", plan.Status)
	}
	if plan.Consumed != 10 {
		t.Errorf("consumed: got %v, want 10", plan.Consumed)
	}
	if len(plan.Blocks) != 1 {
		t.Errorf("blocks: got %d, want 1", len(plan.Blocks))
	}
}

func TestText_PlanLayout_MultiLine(t *testing.T) {
	// "a b c d e" at fontSize 10 → each word ~5pt, maxWidth 12pt → ~1 word per line = 5 lines.
	txt := &Text{
		Content: "a b c d e",
		Style:   Style{FontSize: 10, LineHeight: 1.0},
		Fonts:   &MockFontResolver{},
	}
	plan := txt.PlanLayout(Area{Width: 12, Height: 200})
	if plan.Status != Full {
		t.Errorf("status: got %v, want Full", plan.Status)
	}
	if len(plan.Blocks) == 0 {
		t.Error("expected blocks, got none")
	}
}

func TestText_PlanLayout_Overflow(t *testing.T) {
	// "line one line two line three" at fontSize 20, lineHeight 1.0.
	// Each word "line" = 4 chars * 20 * 0.5 = 40pt. "line one" = 8+1+3=12chars*10=120pt.
	// maxWidth = 100pt → "line" fits, "line one" may or may not fit.
	// Ensure we get overflow when height is limited.
	txt := &Text{
		Content: "alpha beta gamma delta epsilon",
		Style:   Style{FontSize: 20, LineHeight: 1.0},
		Fonts:   &MockFontResolver{},
	}
	// Height = 20pt = 1 line height. The text will wrap to many lines.
	plan := txt.PlanLayout(Area{Width: 200, Height: 20})
	// Depending on wrapping, we may get Full (all on 1 line) or Partial.
	// Force overflow by using a very narrow width.
	txt2 := &Text{
		Content: "alpha beta gamma delta epsilon",
		Style:   Style{FontSize: 20, LineHeight: 1.0},
		Fonts:   &MockFontResolver{},
	}
	plan2 := txt2.PlanLayout(Area{Width: 60, Height: 20})
	// At width 60, "alpha" = 5*10 = 50pt, "alpha beta" = 10*10 = 100 > 60 → wraps.
	// Height 20 = 1 line. Remaining lines overflow.
	_ = plan
	if plan2.Status == Full && strings.Count(txt2.Content, " ") > 0 {
		// Only valid if all text fits on 1 line with the given width.
		// Our content has 4 spaces → at least 5 words.
		// "alpha" = 50pt fits within 60pt width. "alpha beta" = 100pt > 60.
		// So at least 5 lines, but only 1 fits in 20pt.
		// Thus it MUST be Partial (or Nothing if no lines fit).
		// Nothing is returned only if the first line also doesn't fit.
		t.Errorf("expected Partial or Nothing for overflowing text, got Full")
	}
}

func TestText_PlanLayout_NothingWhenNoLineFits(t *testing.T) {
	// Height 0 → nothing fits.
	txt := &Text{
		Content: "hello",
		Style:   Style{FontSize: 10, LineHeight: 1.0},
		Fonts:   &MockFontResolver{},
	}
	plan := txt.PlanLayout(Area{Width: 200, Height: 0})
	if plan.Status != Nothing {
		t.Errorf("expected Nothing for zero height, got %v", plan.Status)
	}
}

func TestText_PlanLayout_EmptyContent(t *testing.T) {
	txt := &Text{
		Content: "",
		Style:   Style{FontSize: 12, LineHeight: 1.0},
		Fonts:   &MockFontResolver{},
	}
	plan := txt.PlanLayout(Area{Width: 200, Height: 100})
	// Empty text: MockFontResolver.LineBreak returns [""], which produces 1 block.
	// The block height = fontSize * lineHeight = 12.
	_ = plan // Just ensure it doesn't panic.
}

func TestText_PlanLayout_OverflowCarriesRemainingText(t *testing.T) {
	// Force multi-line scenario and verify overflow text is not empty.
	txt := &Text{
		Content: "word1 word2 word3 word4 word5",
		Style:   Style{FontSize: 10, LineHeight: 1.0},
		Fonts:   &MockFontResolver{},
	}
	// Width 30 → "word1" = 5*5 = 25pt fits, "word1 word2" = 55pt > 30 → wraps.
	// Height 10 → 1 line fits.
	plan := txt.PlanLayout(Area{Width: 30, Height: 10})
	if plan.Status == Partial {
		if plan.Overflow == nil {
			t.Error("Partial plan must have non-nil Overflow")
		}
		ovTxt, ok := plan.Overflow.(*Text)
		if !ok {
			t.Errorf("overflow should be *Text, got %T", plan.Overflow)
		}
		if ovTxt.Content == "" {
			t.Error("overflow text content should not be empty")
		}
	}
}

func TestText_PlanLayout_Alignment(t *testing.T) {
	tests := []struct {
		name  string
		align Align
	}{
		{"Left", AlignLeft},
		{"Center", AlignCenter},
		{"Right", AlignRight},
		{"Justify", AlignJustify},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			txt := &Text{
				Content: "hello world foo bar",
				Style:   Style{FontSize: 10, LineHeight: 1.0, TextAlign: tc.align},
				Fonts:   &MockFontResolver{},
			}
			plan := txt.PlanLayout(Area{Width: 200, Height: 200})
			if plan.Status == Nothing {
				t.Errorf("alignment %v: got Nothing unexpectedly", tc.name)
			}
		})
	}
}

// ------- Text Measurable tests -------

func TestText_MinWidth(t *testing.T) {
	txt := &Text{
		Content: "short very-long-word end",
		Style:   Style{FontSize: 10},
		Fonts:   &MockFontResolver{},
	}
	min := txt.MinWidth()
	// Longest word is "very-long-word" = 14 chars * 5 = 70.
	want := float64(14) * 10 * 0.5
	if min != want {
		t.Errorf("MinWidth = %v, want %v", min, want)
	}
}

func TestText_MaxWidth(t *testing.T) {
	txt := &Text{
		Content: "hello world",
		Style:   Style{FontSize: 10},
		Fonts:   &MockFontResolver{},
	}
	max := txt.MaxWidth()
	// "hello world" = 11 chars * 5 = 55
	want := float64(11) * 10 * 0.5
	if max != want {
		t.Errorf("MaxWidth = %v, want %v", max, want)
	}
}

// ------- computeTextX tests -------

func TestComputeTextX(t *testing.T) {
	tests := []struct {
		name       string
		align      Align
		lineWidth  float64
		availWidth float64
		line       string
		isLast     bool
		wantX      float64
		wantWS     float64
	}{
		{
			name: "Left", align: AlignLeft,
			lineWidth: 50, availWidth: 100, line: "hello", isLast: false,
			wantX: 0, wantWS: 0,
		},
		{
			name: "Center", align: AlignCenter,
			lineWidth: 60, availWidth: 100, line: "hello", isLast: false,
			wantX: 20, wantWS: 0,
		},
		{
			name: "Right", align: AlignRight,
			lineWidth: 70, availWidth: 100, line: "hello", isLast: false,
			wantX: 30, wantWS: 0,
		},
		{
			name: "Justify mid-line", align: AlignJustify,
			lineWidth: 80, availWidth: 100, line: "hello world", isLast: false,
			wantX: 0, wantWS: 20.0, // (100-80)/1 space = 20
		},
		{
			name: "Justify last line → left", align: AlignJustify,
			lineWidth: 80, availWidth: 100, line: "hello world", isLast: true,
			wantX: 0, wantWS: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x, ws := computeTextX(tc.align, tc.lineWidth, tc.availWidth, tc.line, tc.isLast)
			if x != tc.wantX {
				t.Errorf("x: got %v, want %v", x, tc.wantX)
			}
			if ws != tc.wantWS {
				t.Errorf("wordSpacing: got %v, want %v", ws, tc.wantWS)
			}
		})
	}
}

// ------- Draw closure test -------

func TestText_DrawClosure(t *testing.T) {
	txt := &Text{
		Content: "test draw",
		Style:   Style{FontSize: 12, LineHeight: 1.0, Color: Black},
		Fonts:   &MockFontResolver{},
	}
	plan := txt.PlanLayout(Area{Width: 200, Height: 100})
	if len(plan.Blocks) == 0 {
		t.Fatal("expected blocks")
	}
	// Each block should have a non-nil Draw closure.
	for i, b := range plan.Blocks {
		if b.Draw == nil {
			t.Errorf("block[%d]: Draw closure is nil", i)
		}
	}

	// Invoke Draw with a recording renderer to verify it calls DrawText.
	recorder := &recordingRenderer{}
	for _, b := range plan.Blocks {
		b.Draw(recorder)
	}
	if recorder.drawTextCount == 0 {
		t.Error("expected DrawText to be called at least once")
	}
}

// recordingRenderer counts calls to Renderer methods for testing.
type recordingRenderer struct {
	drawTextCount int
	drawRectCount int
	drawLineCount int
}

func (r *recordingRenderer) DrawText(_ string, _, _ float64, _ FontRef, _ float64, _ Color, _ TextDrawOptions) {
	r.drawTextCount++
}
func (r *recordingRenderer) DrawRect(_, _, _, _ float64, _ *Color, _ *Color, _ float64) {
	r.drawRectCount++
}
func (r *recordingRenderer) DrawLine(_, _, _, _ float64, _ Color, _ float64) {
	r.drawLineCount++
}
func (r *recordingRenderer) DrawImage(_ []byte, _, _, _, _ float64) {}
func (r *recordingRenderer) PushState()                             {}
func (r *recordingRenderer) PopState()                              {}
func (r *recordingRenderer) SetClipRect(_, _, _, _ float64)         {}
