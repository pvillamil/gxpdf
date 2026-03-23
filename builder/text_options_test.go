package builder

import (
	"testing"

	"github.com/coregx/gxpdf/layout"
)

func TestBold(t *testing.T) {
	s := layout.DefaultStyle()
	Bold()(&s)
	if !s.Bold {
		t.Error("Bold() should set Style.Bold = true")
	}
	if s.Font.Weight != layout.WeightBold {
		t.Error("Bold() should set Font.Weight = WeightBold")
	}
}

func TestItalic(t *testing.T) {
	s := layout.DefaultStyle()
	Italic()(&s)
	if !s.Italic {
		t.Error("Italic() should set Style.Italic = true")
	}
	if s.Font.Style != layout.StyleItalic {
		t.Error("Italic() should set Font.Style = StyleItalic")
	}
}

func TestFontSize(t *testing.T) {
	tests := []struct {
		size float64
	}{
		{8}, {10}, {12}, {14}, {18}, {24}, {36}, {72},
	}
	for _, tc := range tests {
		s := layout.DefaultStyle()
		FontSize(tc.size)(&s)
		if s.FontSize != tc.size {
			t.Errorf("FontSize(%f) -> Style.FontSize = %f, want %f", tc.size, s.FontSize, tc.size)
		}
	}
}

func TestFontFamily(t *testing.T) {
	s := layout.DefaultStyle()
	FontFamily("Inter")(&s)
	if s.Font.Family != "Inter" {
		t.Errorf("FontFamily(Inter) -> Font.Family = %q, want Inter", s.Font.Family)
	}
}

func TestTextColor(t *testing.T) {
	s := layout.DefaultStyle()
	TextColor(Red)(&s)
	want := Red.toLayout()
	if s.Color != want {
		t.Errorf("TextColor(Red) -> Style.Color = %v, want %v", s.Color, want)
	}
}

func TestBgColor(t *testing.T) {
	s := layout.DefaultStyle()
	BgColor(LightGray)(&s)
	if s.Background == nil {
		t.Fatal("BgColor should set a non-nil Background")
	}
	want := LightGray.toLayout()
	if *s.Background != want {
		t.Errorf("BgColor(LightGray) -> *Background = %v, want %v", *s.Background, want)
	}
}

func TestAlignment(t *testing.T) {
	tests := []struct {
		name string
		opt  TextOption
		want layout.Align
	}{
		{"AlignLeft", AlignLeft(), layout.AlignLeft},
		{"AlignCenter", AlignCenter(), layout.AlignCenter},
		{"AlignRight", AlignRight(), layout.AlignRight},
		{"AlignJustify", AlignJustify(), layout.AlignJustify},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := layout.DefaultStyle()
			tc.opt(&s)
			if s.TextAlign != tc.want {
				t.Errorf("%s -> TextAlign = %d, want %d", tc.name, s.TextAlign, tc.want)
			}
		})
	}
}

func TestLineHeight(t *testing.T) {
	s := layout.DefaultStyle()
	LineHeight(1.6)(&s)
	if s.LineHeight != 1.6 {
		t.Errorf("LineHeight(1.6) -> Style.LineHeight = %f, want 1.6", s.LineHeight)
	}
}

func TestUnderline(t *testing.T) {
	s := layout.DefaultStyle()
	Underline()(&s)
	if !s.Underline {
		t.Error("Underline() should set Style.Underline = true")
	}
}

func TestStrikethrough(t *testing.T) {
	s := layout.DefaultStyle()
	Strikethrough()(&s)
	if !s.Strikethrough {
		t.Error("Strikethrough() should set Style.Strikethrough = true")
	}
}

func TestLetterSpacing(t *testing.T) {
	s := layout.DefaultStyle()
	LetterSpacing(2.5)(&s)
	if s.LetterSpacing != 2.5 {
		t.Errorf("LetterSpacing(2.5) -> Style.LetterSpacing = %f, want 2.5", s.LetterSpacing)
	}
}

func TestApplyTextOptions_Composition(t *testing.T) {
	base := layout.DefaultStyle()
	result := applyTextOptions(base, []TextOption{
		Bold(),
		FontSize(18),
		TextColor(Red),
		AlignRight(),
		LineHeight(1.5),
	})

	if !result.Bold {
		t.Error("expected Bold = true")
	}
	if result.FontSize != 18 {
		t.Errorf("expected FontSize = 18, got %f", result.FontSize)
	}
	want := Red.toLayout()
	if result.Color != want {
		t.Errorf("expected Color = Red (layout), got %v", result.Color)
	}
	if result.TextAlign != layout.AlignRight {
		t.Errorf("expected TextAlign = AlignRight, got %d", result.TextAlign)
	}
	if result.LineHeight != 1.5 {
		t.Errorf("expected LineHeight = 1.5, got %f", result.LineHeight)
	}
}

func TestApplyTextOptions_LaterOverridesEarlier(t *testing.T) {
	base := layout.DefaultStyle()
	result := applyTextOptions(base, []TextOption{
		FontSize(12),
		FontSize(24), // should win
	})
	if result.FontSize != 24 {
		t.Errorf("expected later FontSize(24) to override earlier, got %f", result.FontSize)
	}
}

func TestApplyTextOptions_BaseNotMutated(t *testing.T) {
	base := layout.DefaultStyle()
	originalSize := base.FontSize

	_ = applyTextOptions(base, []TextOption{FontSize(99)})

	if base.FontSize != originalSize {
		t.Error("applyTextOptions must not mutate the base style")
	}
}

func TestRowHeight(t *testing.T) {
	cfg := applyRowOptions([]RowOption{RowHeight(Mm(30))})
	if cfg.height == nil {
		t.Fatal("RowHeight should set height")
	}
	if cfg.height.Amount != 30 || cfg.height.Unit != layout.UnitMm {
		t.Errorf("RowHeight(30mm) -> height = %v", cfg.height)
	}
}

func TestRowBg(t *testing.T) {
	cfg := applyRowOptions([]RowOption{RowBg(Navy)})
	if cfg.bgColor == nil {
		t.Fatal("RowBg should set bgColor")
	}
	want := Navy.toLayout()
	if *cfg.bgColor != want {
		t.Errorf("RowBg(Navy) -> bgColor = %v, want %v", *cfg.bgColor, want)
	}
}

func TestApplyRowOptions_Empty(t *testing.T) {
	cfg := applyRowOptions(nil)
	if cfg.height != nil {
		t.Error("empty options: height should be nil")
	}
	if cfg.bgColor != nil {
		t.Error("empty options: bgColor should be nil")
	}
}

func TestFitWidth(t *testing.T) {
	cfg := applyImageOptions([]ImageOption{FitWidth(Mm(60))})
	if cfg.width == nil {
		t.Fatal("FitWidth should set width")
	}
	if cfg.width.Amount != 60 || cfg.width.Unit != layout.UnitMm {
		t.Errorf("FitWidth(60mm) -> width = %v", cfg.width)
	}
}

func TestFitHeight(t *testing.T) {
	cfg := applyImageOptions([]ImageOption{FitHeight(Mm(40))})
	if cfg.height == nil {
		t.Fatal("FitHeight should set height")
	}
	if cfg.height.Amount != 40 || cfg.height.Unit != layout.UnitMm {
		t.Errorf("FitHeight(40mm) -> height = %v", cfg.height)
	}
}

func TestLineColorOption(t *testing.T) {
	cfg := applyLineOptions([]LineOption{LineColor(Blue)})
	if cfg.color == nil {
		t.Fatal("LineColor should set color")
	}
	want := Blue.toLayout()
	if *cfg.color != want {
		t.Errorf("LineColor(Blue) -> color = %v, want %v", *cfg.color, want)
	}
}

func TestLineWidthOption(t *testing.T) {
	cfg := applyLineOptions([]LineOption{LineWidth(2.5)})
	if cfg.width != 2.5 {
		t.Errorf("LineWidth(2.5) -> width = %f, want 2.5", cfg.width)
	}
}

func TestApplyLineOptions_DefaultWidth(t *testing.T) {
	cfg := applyLineOptions(nil)
	if cfg.width != 1.0 {
		t.Errorf("default line width should be 1.0, got %f", cfg.width)
	}
}
