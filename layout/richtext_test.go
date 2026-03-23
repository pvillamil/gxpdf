package layout

import (
	"testing"
)

// --- helpers shared across richtext tests ---

// makeFragment constructs a RichTextFragment with a simple style.
func makeFragment(text string, fontSize float64, bold bool) RichTextFragment {
	return RichTextFragment{
		Text: text,
		Style: Style{
			FontSize: fontSize,
			Bold:     bold,
		},
	}
}

// richRecorder is a recording renderer that captures DrawText calls.
type richRecorder struct {
	calls []richDrawCall
}

type richDrawCall struct {
	text     string
	x, y     float64
	fontSize float64
}

func (r *richRecorder) DrawText(text string, x, y float64, _ FontRef, size float64, _ Color, _ TextDrawOptions) {
	r.calls = append(r.calls, richDrawCall{text: text, x: x, y: y, fontSize: size})
}
func (r *richRecorder) DrawRect(_, _, _, _ float64, _ *Color, _ *Color, _ float64) {}
func (r *richRecorder) DrawLine(_, _, _, _ float64, _ Color, _ float64)            {}
func (r *richRecorder) DrawImage(_ []byte, _, _, _, _ float64)                     {}
func (r *richRecorder) PushState()                                                 {}
func (r *richRecorder) PopState()                                                  {}
func (r *richRecorder) SetClipRect(_, _, _, _ float64)                             {}

// --- splitIntoWordsAndSpaces tests ---

func TestSplitIntoWordsAndSpaces_Basic(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"hello", []string{"hello"}},
		{"hello world", []string{"hello", " ", "world"}},
		{"a  b", []string{"a", "  ", "b"}},
		{"  leading", []string{"  ", "leading"}},
		{"trailing  ", []string{"trailing", "  "}},
		{"", nil},
	}
	for _, tc := range tests {
		got := splitIntoWordsAndSpaces(tc.input)
		if len(got) != len(tc.want) {
			t.Errorf("splitIntoWordsAndSpaces(%q) = %v, want %v", tc.input, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("splitIntoWordsAndSpaces(%q)[%d] = %q, want %q", tc.input, i, got[i], tc.want[i])
			}
		}
	}
}

func TestIsAllSpaces(t *testing.T) {
	if !isAllSpaces("   ") {
		t.Error("expected true for all spaces")
	}
	if isAllSpaces("") {
		t.Error("expected false for empty string")
	}
	if isAllSpaces("a b") {
		t.Error("expected false for mixed")
	}
}

// --- RichText.PlanLayout tests ---

func TestRichText_PlanLayout_Empty(t *testing.T) {
	rt := &RichText{
		Fragments: nil,
		Fonts:     &MockFontResolver{},
	}
	plan := rt.PlanLayout(Area{Width: 200, Height: 100})
	if plan.Status != Full {
		t.Errorf("empty RichText: status = %v, want Full", plan.Status)
	}
	if plan.Consumed != 0 {
		t.Errorf("empty RichText: consumed = %v, want 0", plan.Consumed)
	}
	if len(plan.Blocks) != 0 {
		t.Errorf("empty RichText: blocks = %d, want 0", len(plan.Blocks))
	}
}

func TestRichText_PlanLayout_EmptyTextFragment(t *testing.T) {
	rt := &RichText{
		Fragments: []RichTextFragment{{Text: ""}},
		Fonts:     &MockFontResolver{},
	}
	plan := rt.PlanLayout(Area{Width: 200, Height: 100})
	if plan.Status != Full {
		t.Errorf("empty fragment: status = %v, want Full", plan.Status)
	}
}

func TestRichText_PlanLayout_SingleFragment_SingleLine(t *testing.T) {
	// "Hi" at fontSize 10, lineHeight 1.0 → lineSpacing = 10, consumed = 10.
	rt := &RichText{
		Fragments:  []RichTextFragment{makeFragment("Hi", 10, false)},
		LineHeight: 1.0,
		Fonts:      &MockFontResolver{},
	}
	plan := rt.PlanLayout(Area{Width: 200, Height: 100})

	if plan.Status != Full {
		t.Errorf("status = %v, want Full", plan.Status)
	}
	if plan.Consumed != 10 {
		t.Errorf("consumed = %v, want 10", plan.Consumed)
	}
	if len(plan.Blocks) != 1 {
		t.Fatalf("blocks = %d, want 1", len(plan.Blocks))
	}
	if plan.Blocks[0].Draw == nil {
		t.Error("block Draw closure is nil")
	}
}

func TestRichText_PlanLayout_SingleFragment_BehavesLikeText(t *testing.T) {
	// A RichText with one fragment at size 12, lineHeight 1.2 should produce
	// the same vertical consumption as a Text element with the same content.
	content := "Hello world test content"
	rt := &RichText{
		Fragments:  []RichTextFragment{makeFragment(content, 12, false)},
		LineHeight: 1.2,
		Fonts:      &MockFontResolver{},
	}
	txt := &Text{
		Content: content,
		Style:   Style{FontSize: 12, LineHeight: 1.2},
		Fonts:   &MockFontResolver{},
	}
	area := Area{Width: 100, Height: 500}

	rtPlan := rt.PlanLayout(area)
	txtPlan := txt.PlanLayout(area)

	if rtPlan.Status != Full || txtPlan.Status != Full {
		t.Skip("both should be Full for this test to be valid")
	}
	// Same number of lines.
	if len(rtPlan.Blocks) != len(txtPlan.Blocks) {
		t.Errorf("line count: RichText=%d Text=%d", len(rtPlan.Blocks), len(txtPlan.Blocks))
	}
	// Same total consumed height.
	if abs(rtPlan.Consumed-txtPlan.Consumed) > 0.1 {
		t.Errorf("consumed: RichText=%v Text=%v", rtPlan.Consumed, txtPlan.Consumed)
	}
}

func TestRichText_PlanLayout_TwoFragmentsDifferentStyles(t *testing.T) {
	// "Normal " (size 10) + "Bold" (size 10, bold) — same size, so single line.
	rt := &RichText{
		Fragments: []RichTextFragment{
			makeFragment("Normal ", 10, false),
			makeFragment("Bold", 10, true),
		},
		LineHeight: 1.0,
		Fonts:      &MockFontResolver{},
	}
	plan := rt.PlanLayout(Area{Width: 200, Height: 100})

	if plan.Status != Full {
		t.Errorf("status = %v, want Full", plan.Status)
	}
	if len(plan.Blocks) != 1 {
		t.Errorf("blocks = %d, want 1 (all fits on one line)", len(plan.Blocks))
	}
	// Draw the block and verify two DrawText calls (one per word-run).
	rec := &richRecorder{}
	plan.Blocks[0].Draw(rec)
	if len(rec.calls) < 2 {
		t.Errorf("DrawText calls = %d, want >= 2 (Normal + Bold)", len(rec.calls))
	}
}

func TestRichText_PlanLayout_MixedFontSizes_BaselineAlignment(t *testing.T) {
	// Large text (size 20) and small text (size 10) on the same line.
	// The line box height = 20 * lineHeight.
	// Small text y offset = halfLeading + (20 - 10) = halfLeading + 10.
	// Large text y offset = halfLeading + 0 = halfLeading.
	rt := &RichText{
		Fragments: []RichTextFragment{
			makeFragment("Big ", 20, false),
			makeFragment("small", 10, false),
		},
		LineHeight: 1.0,
		Fonts:      &MockFontResolver{},
	}
	plan := rt.PlanLayout(Area{Width: 500, Height: 100})

	if plan.Status != Full {
		t.Errorf("status = %v, want Full", plan.Status)
	}
	// Line height is driven by the tallest run: 20 * 1.0 = 20.
	if abs(plan.Consumed-20) > 0.01 {
		t.Errorf("consumed = %v, want 20 (driven by max font size)", plan.Consumed)
	}

	rec := &richRecorder{}
	plan.Blocks[0].Draw(rec)
	if len(rec.calls) < 2 {
		t.Fatalf("expected >= 2 DrawText calls, got %d", len(rec.calls))
	}

	// Find the large and small text calls.
	var bigY, smallY float64
	for _, c := range rec.calls {
		if c.fontSize == 20 {
			bigY = c.y
		}
		if c.fontSize == 10 {
			smallY = c.y
		}
	}
	// Small text must be below (larger Y) big text for baseline alignment.
	if smallY <= bigY {
		t.Errorf("small text Y (%v) should be > big text Y (%v) for baseline alignment", smallY, bigY)
	}
}

func TestRichText_PlanLayout_LineWrapping(t *testing.T) {
	// "word " repeated: force wrapping by using narrow width.
	// Each "word" = 4 chars * 10 * 0.5 = 20pt at fontSize 10.
	// availWidth = 50pt → "word word" = 45pt fits, "word word word" = 70pt > 50.
	rt := &RichText{
		Fragments: []RichTextFragment{
			makeFragment("word word word word word", 10, false),
		},
		LineHeight: 1.0,
		Fonts:      &MockFontResolver{},
	}
	plan := rt.PlanLayout(Area{Width: 50, Height: 500})

	if plan.Status != Full {
		t.Errorf("status = %v, want Full", plan.Status)
	}
	if len(plan.Blocks) < 2 {
		t.Errorf("expected >= 2 lines for narrow width, got %d", len(plan.Blocks))
	}
}

func TestRichText_PlanLayout_WrapAcrossFragmentBoundary(t *testing.T) {
	// "Hello " in fragment 1, "world foo bar" in fragment 2.
	// At width 60pt (size 10): "Hello" = 25pt, "Hello world" = 55pt fits,
	// "Hello world foo" = 75pt > 60 → wraps after "world".
	rt := &RichText{
		Fragments: []RichTextFragment{
			makeFragment("Hello ", 10, false),
			makeFragment("world foo bar", 10, false),
		},
		LineHeight: 1.0,
		Fonts:      &MockFontResolver{},
	}
	plan := rt.PlanLayout(Area{Width: 60, Height: 500})

	if plan.Status != Full {
		t.Errorf("status = %v, want Full", plan.Status)
	}
	if len(plan.Blocks) < 2 {
		t.Errorf("expected >= 2 lines, got %d", len(plan.Blocks))
	}
}

func TestRichText_PlanLayout_Overflow_Partial(t *testing.T) {
	// Many words, narrow width, limited height → Partial with Overflow.
	rt := &RichText{
		Fragments: []RichTextFragment{
			makeFragment("one two three four five six seven eight", 10, false),
		},
		LineHeight: 1.0,
		Fonts:      &MockFontResolver{},
	}
	// Width 40: each "word" ~ 2-5 chars * 5pt. Height 10: only 1 line fits.
	plan := rt.PlanLayout(Area{Width: 40, Height: 10})

	if plan.Status != Partial && plan.Status != Nothing {
		t.Errorf("expected Partial or Nothing for limited height, got %v", plan.Status)
	}
	if plan.Status == Partial {
		if plan.Overflow == nil {
			t.Error("Partial plan must have non-nil Overflow")
		}
		ovRT, ok := plan.Overflow.(*RichText)
		if !ok {
			t.Errorf("Overflow should be *RichText, got %T", plan.Overflow)
		}
		if len(ovRT.Fragments) == 0 {
			t.Error("Overflow RichText should have fragments")
		}
	}
}

func TestRichText_PlanLayout_Overflow_Nothing(t *testing.T) {
	// Height = 0 → nothing fits at all.
	rt := &RichText{
		Fragments:  []RichTextFragment{makeFragment("hello", 10, false)},
		LineHeight: 1.0,
		Fonts:      &MockFontResolver{},
	}
	plan := rt.PlanLayout(Area{Width: 200, Height: 0})
	if plan.Status != Nothing {
		t.Errorf("height=0: status = %v, want Nothing", plan.Status)
	}
}

func TestRichText_PlanLayout_OverflowIsRichText(t *testing.T) {
	// Verify that overflow carries the same font resolver and alignment.
	rt := &RichText{
		Fragments: []RichTextFragment{
			makeFragment("alpha beta gamma delta epsilon zeta", 10, false),
		},
		LineHeight: 1.0,
		Align:      AlignCenter,
		Fonts:      &MockFontResolver{},
	}
	plan := rt.PlanLayout(Area{Width: 40, Height: 10})
	if plan.Status == Partial {
		ovRT, ok := plan.Overflow.(*RichText)
		if !ok {
			t.Fatalf("Overflow is %T, want *RichText", plan.Overflow)
		}
		if ovRT.Align != AlignCenter {
			t.Errorf("Overflow.Align = %v, want AlignCenter", ovRT.Align)
		}
		if ovRT.Fonts != rt.Fonts {
			t.Error("Overflow.Fonts should match original Fonts")
		}
	}
}

func TestRichText_PlanLayout_AlignCenter(t *testing.T) {
	// Single word "Hi" (size 10, width 10) in 100pt container.
	// Center: X = (100 - 10) / 2 = 45.
	rt := &RichText{
		Fragments:  []RichTextFragment{makeFragment("Hi", 10, false)},
		LineHeight: 1.0,
		Align:      AlignCenter,
		Fonts:      &MockFontResolver{},
	}
	plan := rt.PlanLayout(Area{Width: 100, Height: 100})
	if len(plan.Blocks) == 0 {
		t.Fatal("expected blocks")
	}
	rec := &richRecorder{}
	plan.Blocks[0].Draw(rec)
	if len(rec.calls) == 0 {
		t.Fatal("expected DrawText calls")
	}
	wantX := (100.0 - float64(2)*10*0.5) / 2 // (100 - 10) / 2 = 45
	if abs(rec.calls[0].x-wantX) > 0.5 {
		t.Errorf("center X = %v, want ~%v", rec.calls[0].x, wantX)
	}
}

func TestRichText_PlanLayout_AlignRight(t *testing.T) {
	// "Hi" (size 10, width 10) right-aligned in 100pt → X = 90.
	rt := &RichText{
		Fragments:  []RichTextFragment{makeFragment("Hi", 10, false)},
		LineHeight: 1.0,
		Align:      AlignRight,
		Fonts:      &MockFontResolver{},
	}
	plan := rt.PlanLayout(Area{Width: 100, Height: 100})
	if len(plan.Blocks) == 0 {
		t.Fatal("expected blocks")
	}
	rec := &richRecorder{}
	plan.Blocks[0].Draw(rec)
	wantX := 100.0 - float64(2)*10*0.5 // 100 - 10 = 90
	if abs(rec.calls[0].x-wantX) > 0.5 {
		t.Errorf("right X = %v, want ~%v", rec.calls[0].x, wantX)
	}
}

func TestRichText_PlanLayout_AlignJustify_LastLineIsLeft(t *testing.T) {
	// The last line of a justified paragraph is left-aligned (X = 0).
	rt := &RichText{
		Fragments: []RichTextFragment{
			// 5 words that wrap onto 2 lines, making the last line a single word.
			makeFragment("one two three four five", 10, false),
		},
		LineHeight: 1.0,
		Align:      AlignJustify,
		Fonts:      &MockFontResolver{},
	}
	// Width 50 → "one two" = 35pt fits, "one two three" = 55pt > 50 → wraps.
	plan := rt.PlanLayout(Area{Width: 50, Height: 500})
	if len(plan.Blocks) < 2 {
		t.Skip("test needs at least 2 lines to verify last-line left-align")
	}
	lastBlock := plan.Blocks[len(plan.Blocks)-1]
	rec := &richRecorder{}
	lastBlock.Draw(rec)
	if len(rec.calls) == 0 {
		t.Fatal("no DrawText calls on last line")
	}
	// Last line first word should start at X=0 (left aligned).
	if rec.calls[0].x != 0 {
		t.Errorf("last line first word X = %v, want 0 (left-aligned last line)", rec.calls[0].x)
	}
}

func TestRichText_PlanLayout_DrawClosures_AllNonNil(t *testing.T) {
	rt := &RichText{
		Fragments: []RichTextFragment{
			makeFragment("First fragment ", 10, false),
			makeFragment("Second fragment", 10, true),
		},
		LineHeight: 1.0,
		Fonts:      &MockFontResolver{},
	}
	plan := rt.PlanLayout(Area{Width: 500, Height: 500})
	for i, b := range plan.Blocks {
		if b.Draw == nil {
			t.Errorf("block[%d]: Draw closure is nil", i)
		}
	}
}

func TestRichText_PlanLayout_DrawCallsAllFragments(t *testing.T) {
	// Verify that Draw on a block containing multiple run styles calls
	// DrawText once per non-space word run.
	rt := &RichText{
		Fragments: []RichTextFragment{
			makeFragment("alpha ", 10, false),
			makeFragment("beta ", 12, true),
			makeFragment("gamma", 8, false),
		},
		LineHeight: 1.0,
		Fonts:      &MockFontResolver{},
	}
	plan := rt.PlanLayout(Area{Width: 500, Height: 500})
	if len(plan.Blocks) == 0 {
		t.Fatal("expected blocks")
	}

	var totalCalls int
	rec := &richRecorder{}
	for _, b := range plan.Blocks {
		b.Draw(rec)
	}
	totalCalls = len(rec.calls)
	// 3 word runs → at least 3 DrawText calls.
	if totalCalls < 3 {
		t.Errorf("total DrawText calls = %d, want >= 3", totalCalls)
	}
}

func TestRichText_PlanLayout_HalfLeading(t *testing.T) {
	// lineHeight 1.4 at fontSize 10 → lineSpacing = 14.
	// halfLeading = (14 - 10) / 2 = 2.
	// The single run Y should equal halfLeading (= 2).
	rt := &RichText{
		Fragments:  []RichTextFragment{makeFragment("Hi", 10, false)},
		LineHeight: 1.4,
		Fonts:      &MockFontResolver{},
	}
	plan := rt.PlanLayout(Area{Width: 200, Height: 100})
	if len(plan.Blocks) == 0 {
		t.Fatal("expected blocks")
	}
	rec := &richRecorder{}
	plan.Blocks[0].Draw(rec)
	if len(rec.calls) == 0 {
		t.Fatal("no DrawText calls")
	}
	wantY := (14.0 - 10.0) / 2 // halfLeading = 2.0
	if abs(rec.calls[0].y-wantY) > 0.01 {
		t.Errorf("halfLeading Y = %v, want %v", rec.calls[0].y, wantY)
	}
}

func TestRichText_PlanLayout_DefaultLineHeight(t *testing.T) {
	// LineHeight 0 should default to 1.2.
	rt := &RichText{
		Fragments:  []RichTextFragment{makeFragment("Hello", 10, false)},
		LineHeight: 0,
		Fonts:      &MockFontResolver{},
	}
	plan := rt.PlanLayout(Area{Width: 200, Height: 100})
	// With default lineHeight 1.2, consumed = 10 * 1.2 = 12.
	if abs(plan.Consumed-12.0) > 0.01 {
		t.Errorf("consumed = %v, want 12 (default lineHeight 1.2)", plan.Consumed)
	}
}

func TestRichText_PlanLayout_NilFontsUseMock(t *testing.T) {
	// Nil Fonts should not panic; it falls back to MockFontResolver.
	rt := &RichText{
		Fragments:  []RichTextFragment{makeFragment("hello", 10, false)},
		LineHeight: 1.0,
		Fonts:      nil,
	}
	// Should not panic.
	plan := rt.PlanLayout(Area{Width: 200, Height: 100})
	if plan.Status != Full {
		t.Errorf("status = %v, want Full", plan.Status)
	}
}

// --- MinWidth / MaxWidth tests ---

func TestRichText_MinWidth(t *testing.T) {
	// "Hello world test" at size 10.
	// Longest word: "Hello" = 5*5 = 25pt; "world" = 5*5 = 25pt; "test" = 4*5 = 20pt.
	rt := &RichText{
		Fragments: []RichTextFragment{
			makeFragment("Hello world test", 10, false),
		},
		Fonts: &MockFontResolver{},
	}
	min := rt.MinWidth()
	want := float64(5) * 10 * 0.5 // 25
	if abs(min-want) > 0.01 {
		t.Errorf("MinWidth = %v, want %v", min, want)
	}
}

func TestRichText_MinWidth_AcrossFragments(t *testing.T) {
	// Fragment 1: "hello" at size 10 → 25pt.
	// Fragment 2: "verylongword" at size 10 → 12*5 = 60pt.
	// MinWidth = 60.
	rt := &RichText{
		Fragments: []RichTextFragment{
			makeFragment("hello", 10, false),
			makeFragment(" verylongword", 10, false),
		},
		Fonts: &MockFontResolver{},
	}
	min := rt.MinWidth()
	want := float64(12) * 10 * 0.5 // "verylongword" = 12 chars = 60
	if abs(min-want) > 0.01 {
		t.Errorf("MinWidth = %v, want %v", min, want)
	}
}

func TestRichText_MaxWidth(t *testing.T) {
	// Two fragments: "hello " (6 chars, size 10) + "world" (5 chars, size 10).
	// MaxWidth = (6 + 5) * 10 * 0.5 = 55.
	rt := &RichText{
		Fragments: []RichTextFragment{
			makeFragment("hello ", 10, false),
			makeFragment("world", 10, false),
		},
		Fonts: &MockFontResolver{},
	}
	max := rt.MaxWidth()
	want := float64(11) * 10 * 0.5 // 55
	if abs(max-want) > 0.01 {
		t.Errorf("MaxWidth = %v, want %v", max, want)
	}
}

func TestRichText_MaxWidth_MixedSizes(t *testing.T) {
	// "Hi" at size 10 (5pt each char × 2 = 10pt) + "Big" at size 20 (10pt × 3 = 30pt).
	// MaxWidth = 10 + 30 = 40.
	rt := &RichText{
		Fragments: []RichTextFragment{
			makeFragment("Hi", 10, false),
			makeFragment("Big", 20, false),
		},
		Fonts: &MockFontResolver{},
	}
	max := rt.MaxWidth()
	want := float64(2)*10*0.5 + float64(3)*20*0.5 // 10 + 30 = 40
	if abs(max-want) > 0.01 {
		t.Errorf("MaxWidth = %v, want %v", max, want)
	}
}

// --- fillRichLines tests ---

func TestFillRichLines_SingleWord(t *testing.T) {
	runs := []richRun{
		{text: "hello", width: 25, fontSize: 10, isSpace: false},
	}
	lines := fillRichLines(runs, 100)
	if len(lines) != 1 {
		t.Errorf("lines = %d, want 1", len(lines))
	}
	if len(lines[0]) != 1 {
		t.Errorf("line[0] runs = %d, want 1", len(lines[0]))
	}
}

func TestFillRichLines_LeadingSpacesSkipped(t *testing.T) {
	runs := []richRun{
		{text: " ", width: 5, fontSize: 10, isSpace: true},
		{text: "hello", width: 25, fontSize: 10, isSpace: false},
	}
	lines := fillRichLines(runs, 100)
	if len(lines) != 1 {
		t.Fatalf("lines = %d, want 1", len(lines))
	}
	// Leading space should be dropped.
	if lines[0][0].isSpace {
		t.Error("leading space run should not be on the line")
	}
}

func TestFillRichLines_TrailingSpacesTrimmed(t *testing.T) {
	runs := []richRun{
		{text: "hello", width: 25, fontSize: 10, isSpace: false},
		{text: " ", width: 5, fontSize: 10, isSpace: true},
	}
	lines := fillRichLines(runs, 100)
	if len(lines) != 1 {
		t.Fatalf("lines = %d, want 1", len(lines))
	}
	last := lines[0][len(lines[0])-1]
	if last.isSpace {
		t.Error("trailing space should be trimmed from line")
	}
}

func TestFillRichLines_WordsWrap(t *testing.T) {
	// "word word word" at width=30pt each word=20pt.
	// Line 1: "word" (20) + " " (5) + "word" would be 45 > 30 → break.
	runs := []richRun{
		{text: "word", width: 20, fontSize: 10, isSpace: false},
		{text: " ", width: 5, fontSize: 10, isSpace: true},
		{text: "word", width: 20, fontSize: 10, isSpace: false},
		{text: " ", width: 5, fontSize: 10, isSpace: true},
		{text: "word", width: 20, fontSize: 10, isSpace: false},
	}
	lines := fillRichLines(runs, 30)
	if len(lines) < 2 {
		t.Errorf("expected >= 2 lines, got %d", len(lines))
	}
}

func TestFillRichLines_EmptyRuns(t *testing.T) {
	lines := fillRichLines(nil, 100)
	if len(lines) != 0 {
		t.Errorf("empty runs: lines = %d, want 0", len(lines))
	}
}

// --- maxRichFontSize tests ---

func TestMaxRichFontSize_Basic(t *testing.T) {
	line := []richRun{
		{fontSize: 10},
		{fontSize: 20},
		{fontSize: 8},
	}
	if got := maxRichFontSize(line); got != 20 {
		t.Errorf("maxRichFontSize = %v, want 20", got)
	}
}

func TestMaxRichFontSize_EmptyLine(t *testing.T) {
	if got := maxRichFontSize(nil); got != 12 {
		t.Errorf("maxRichFontSize(nil) = %v, want 12 (default)", got)
	}
}

// --- rebuildRichOverflow tests ---

func TestRebuildRichOverflow_PreservesAlignment(t *testing.T) {
	s := Style{FontSize: 10}
	s = s.effective()
	lines := [][]richRun{
		{{text: "hello", style: s, fontSize: 10, isSpace: false}},
		{{text: "world", style: s, fontSize: 10, isSpace: false}},
	}
	overflow := rebuildRichOverflow(lines, AlignCenter, 1.5, &MockFontResolver{})
	if overflow.Align != AlignCenter {
		t.Errorf("Align = %v, want AlignCenter", overflow.Align)
	}
	if overflow.LineHeight != 1.5 {
		t.Errorf("LineHeight = %v, want 1.5", overflow.LineHeight)
	}
}

func TestRebuildRichOverflow_FragmentsNonEmpty(t *testing.T) {
	s := Style{FontSize: 10}
	s = s.effective()
	lines := [][]richRun{
		{
			{text: "one", style: s, fontSize: 10, isSpace: false},
			{text: " ", style: s, fontSize: 10, isSpace: true},
			{text: "two", style: s, fontSize: 10, isSpace: false},
		},
	}
	overflow := rebuildRichOverflow(lines, AlignLeft, 1.2, nil)
	if len(overflow.Fragments) == 0 {
		t.Error("overflow fragments should not be empty")
	}
}

// --- Link area tests ---

func TestRichText_PlanLayout_LinkFragment(t *testing.T) {
	rt := &RichText{
		Fragments: []RichTextFragment{
			{
				Text:  "click here",
				Style: Style{FontSize: 10},
				URL:   "https://example.com",
			},
		},
		LineHeight: 1.0,
		Fonts:      &MockFontResolver{},
	}
	plan := rt.PlanLayout(Area{Width: 200, Height: 100})
	if len(plan.Blocks) == 0 {
		t.Fatal("expected blocks")
	}
	// The block should carry link areas for the URL runs.
	foundLink := false
	for _, b := range plan.Blocks {
		for _, l := range b.Links {
			if l.URL == "https://example.com" {
				foundLink = true
			}
		}
	}
	if !foundLink {
		t.Error("expected link area with URL 'https://example.com' in blocks")
	}
}

// --- abs helper ---

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
