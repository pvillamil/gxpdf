package layout

import (
	"math"
	"testing"
)

func TestValueResolve(t *testing.T) {
	tests := []struct {
		name       string
		value      Value
		parentSize float64
		fontSize   float64
		want       float64
		tolerance  float64
	}{
		{name: "Pt", value: Pt(72), parentSize: 500, fontSize: 12, want: 72},
		{name: "Mm 1mm", value: Mm(1), parentSize: 500, fontSize: 12, want: 2.834645669, tolerance: 1e-6},
		{name: "Mm 25.4mm = 1in", value: Mm(25.4), parentSize: 500, fontSize: 12, want: 72, tolerance: 1e-3},
		{name: "Cm 1cm", value: Cm(1), parentSize: 500, fontSize: 12, want: 28.34645669, tolerance: 1e-4},
		{name: "In 1in", value: In(1), parentSize: 500, fontSize: 12, want: 72},
		{name: "In 2in", value: In(2), parentSize: 500, fontSize: 12, want: 144},
		{name: "Pct 50%", value: Pct(50), parentSize: 400, fontSize: 12, want: 200},
		{name: "Pct 100%", value: Pct(100), parentSize: 300, fontSize: 12, want: 300},
		{name: "Fr returns 0", value: Fr(1), parentSize: 400, fontSize: 12, want: 0},
		{name: "Auto returns 0", value: Auto(), parentSize: 400, fontSize: 12, want: 0},
		{name: "Pt zero", value: Pt(0), parentSize: 500, fontSize: 12, want: 0},
		{name: "Mm zero", value: Mm(0), parentSize: 500, fontSize: 12, want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.value.Resolve(tc.parentSize, tc.fontSize)
			tol := tc.tolerance
			if tol == 0 {
				tol = 1e-9
			}
			if math.Abs(got-tc.want) > tol {
				t.Errorf("Resolve() = %v, want %v (tolerance %v)", got, tc.want, tol)
			}
		})
	}
}

func TestValueIsAuto(t *testing.T) {
	if !Auto().IsAuto() {
		t.Error("Auto().IsAuto() should be true")
	}
	if Pt(10).IsAuto() {
		t.Error("Pt(10).IsAuto() should be false")
	}
	if Mm(5).IsAuto() {
		t.Error("Mm(5).IsAuto() should be false")
	}
}

func TestEdgesResolve(t *testing.T) {
	edges := Edges{
		Top:    Mm(10),
		Right:  Pt(20),
		Bottom: Pct(5),
		Left:   In(0.5),
	}
	resolved := edges.Resolve(400, 600, 12)

	wantTop := 10 * ptPerMm
	wantRight := 20.0
	wantBottom := 5.0 / 100.0 * 600
	wantLeft := 0.5 * 72

	tolerance := 1e-4

	if math.Abs(resolved.Top-wantTop) > tolerance {
		t.Errorf("Top: got %v, want %v", resolved.Top, wantTop)
	}
	if math.Abs(resolved.Right-wantRight) > tolerance {
		t.Errorf("Right: got %v, want %v", resolved.Right, wantRight)
	}
	if math.Abs(resolved.Bottom-wantBottom) > tolerance {
		t.Errorf("Bottom: got %v, want %v", resolved.Bottom, wantBottom)
	}
	if math.Abs(resolved.Left-wantLeft) > tolerance {
		t.Errorf("Left: got %v, want %v", resolved.Left, wantLeft)
	}
}

func TestResolvedEdgesHorizontalVertical(t *testing.T) {
	r := ResolvedEdges{Top: 10, Right: 20, Bottom: 15, Left: 25}
	if got := r.Horizontal(); got != 45 {
		t.Errorf("Horizontal() = %v, want 45", got)
	}
	if got := r.Vertical(); got != 25 {
		t.Errorf("Vertical() = %v, want 25", got)
	}
}

func TestUniformEdges(t *testing.T) {
	edges := UniformEdges(Pt(10))
	resolved := edges.Resolve(200, 200, 12)
	if resolved.Top != 10 || resolved.Right != 10 || resolved.Bottom != 10 || resolved.Left != 10 {
		t.Errorf("UniformEdges: got %+v, want all 10", resolved)
	}
}

func TestPageSizes(t *testing.T) {
	// A4 is approximately 210mm × 297mm
	wantW := 210 * ptPerMm
	wantH := 297 * ptPerMm
	tol := 1.0 // 1 point tolerance for rounding

	if math.Abs(PageA4.Width-wantW) > tol {
		t.Errorf("A4 width: got %v, want ~%v", PageA4.Width, wantW)
	}
	if math.Abs(PageA4.Height-wantH) > tol {
		t.Errorf("A4 height: got %v, want ~%v", PageA4.Height, wantH)
	}
}

func TestConstructors(t *testing.T) {
	tests := []struct {
		name string
		v    Value
		unit Unit
	}{
		{"Pt", Pt(5), UnitPt},
		{"Mm", Mm(5), UnitMm},
		{"Cm", Cm(5), UnitCm},
		{"In", In(5), UnitIn},
		{"Pct", Pct(50), UnitPct},
		{"Fr", Fr(2), UnitFr},
		{"Auto", Auto(), UnitAuto},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.v.Unit != tc.unit {
				t.Errorf("unit: got %v, want %v", tc.v.Unit, tc.unit)
			}
		})
	}
}
