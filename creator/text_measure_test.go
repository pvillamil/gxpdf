package creator

import (
	"testing"
)

func TestMeasureText(t *testing.T) {
	tests := []struct {
		font FontName
		text string
		size float64
	}{
		{Helvetica, "Hello", 12},
		{HelveticaBold, "Hello", 12},
		{TimesRoman, "Hello", 12},
		{Courier, "Hello", 12},
	}

	for _, tt := range tests {
		t.Run(string(tt.font), func(t *testing.T) {
			w := MeasureText(tt.font, tt.text, tt.size)
			if w <= 0 {
				t.Errorf("MeasureText(%s, %q, %g) = %g, want > 0", tt.font, tt.text, tt.size, w)
			}
		})
	}
}

func TestMeasureTextEmpty(t *testing.T) {
	w := MeasureText(Helvetica, "", 12)
	if w != 0 {
		t.Errorf("MeasureText(Helvetica, \"\", 12) = %g, want 0", w)
	}
}

func TestMeasureTextUnknownFont(t *testing.T) {
	w := MeasureText(FontName("UnknownFont"), "Hello", 12)
	if w != 0 {
		t.Errorf("MeasureText(UnknownFont, ...) = %g, want 0", w)
	}
}

func TestMeasureTextCourier(t *testing.T) {
	// Courier is monospaced: all characters should have equal width.
	w1 := MeasureText(Courier, "iiiii", 12)
	w2 := MeasureText(Courier, "MMMMM", 12)
	if w1 != w2 {
		t.Errorf("Courier should be monospaced: 'iiiii'=%g, 'MMMMM'=%g", w1, w2)
	}
}

func TestFontAscender(t *testing.T) {
	a := FontAscender(Helvetica, 12)
	if a <= 0 {
		t.Errorf("FontAscender(Helvetica, 12) = %g, want > 0", a)
	}
	// Helvetica ascender = 718 font units, at 12pt: 718*12/1000 = 8.616
	expected := 718.0 * 12.0 / 1000.0
	if a != expected {
		t.Errorf("FontAscender(Helvetica, 12) = %g, want %g", a, expected)
	}
}

func TestFontDescender(t *testing.T) {
	d := FontDescender(Helvetica, 12)
	if d >= 0 {
		t.Errorf("FontDescender(Helvetica, 12) = %g, want < 0", d)
	}
	// Helvetica descender = -207 font units, at 12pt: -207*12/1000 = -2.484
	expected := -207.0 * 12.0 / 1000.0
	if d != expected {
		t.Errorf("FontDescender(Helvetica, 12) = %g, want %g", d, expected)
	}
}

func TestFontCapHeight(t *testing.T) {
	h := FontCapHeight(Helvetica, 12)
	if h <= 0 {
		t.Errorf("FontCapHeight(Helvetica, 12) = %g, want > 0", h)
	}
}

func TestFontLineHeight(t *testing.T) {
	lh := FontLineHeight(Helvetica, 12)
	if lh <= 0 {
		t.Errorf("FontLineHeight(Helvetica, 12) = %g, want > 0", lh)
	}
	// Line height = (718 - (-207)) * 12 / 1000 = 925 * 12 / 1000 = 11.1
	expected := 925.0 * 12.0 / 1000.0
	if lh != expected {
		t.Errorf("FontLineHeight(Helvetica, 12) = %g, want %g", lh, expected)
	}
}

func TestFontMetricsUnknownFont(t *testing.T) {
	if FontAscender(FontName("Unknown"), 12) != 0 {
		t.Error("FontAscender should return 0 for unknown font")
	}
	if FontDescender(FontName("Unknown"), 12) != 0 {
		t.Error("FontDescender should return 0 for unknown font")
	}
	if FontCapHeight(FontName("Unknown"), 12) != 0 {
		t.Error("FontCapHeight should return 0 for unknown font")
	}
	if FontLineHeight(FontName("Unknown"), 12) != 0 {
		t.Error("FontLineHeight should return 0 for unknown font")
	}
}

func TestFontLineHeightRelation(t *testing.T) {
	// LineHeight should equal Ascender - Descender
	size := 12.0
	lh := FontLineHeight(Helvetica, size)
	a := FontAscender(Helvetica, size)
	d := FontDescender(Helvetica, size)

	if lh != a-d {
		t.Errorf("LineHeight (%g) != Ascender (%g) - Descender (%g) = %g", lh, a, d, a-d)
	}
}
