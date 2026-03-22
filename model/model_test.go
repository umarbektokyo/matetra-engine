package model

import (
	"math"
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name       string
		in         Number
		wantValue  float64
		wantBase   int64
	}{
		{"zero", Number{Value: 0, Base: 0}, 0, 0},
		{"already normal", Number{Value: 3.14, Base: 0}, 3.14, 0},
		{"large value", Number{Value: 314.0, Base: 0}, 3.14, 2},
		{"small value", Number{Value: 0.314, Base: 0}, 3.14, -1},
		{"negative", Number{Value: -42.0, Base: 0}, -4.2, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := tt.in
			n.Normalize()
			if math.Abs(n.Value-tt.wantValue) > 0.001 || n.Base != tt.wantBase {
				t.Errorf("got {%f, %d}, want {%f, %d}", n.Value, n.Base, tt.wantValue, tt.wantBase)
			}
		})
	}
}

func TestNumAdd(t *testing.T) {
	a := NumFromFloat(100)
	b := NumFromFloat(200)
	r := NumAdd(a, b)
	if math.Abs(r.ToFloat64()-300) > 0.001 {
		t.Errorf("100 + 200 = %f, want 300", r.ToFloat64())
	}
}

func TestNumSub(t *testing.T) {
	r := NumSub(NumFromFloat(500), NumFromFloat(200))
	if math.Abs(r.ToFloat64()-300) > 0.001 {
		t.Errorf("500 - 200 = %f, want 300", r.ToFloat64())
	}
}

func TestNumMul(t *testing.T) {
	r := NumMul(NumFromFloat(12), NumFromFloat(12))
	if math.Abs(r.ToFloat64()-144) > 0.001 {
		t.Errorf("12 * 12 = %f, want 144", r.ToFloat64())
	}
}

func TestNumDiv(t *testing.T) {
	r, err := NumDiv(NumFromFloat(100), NumFromFloat(4))
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(r.ToFloat64()-25) > 0.001 {
		t.Errorf("100 / 4 = %f, want 25", r.ToFloat64())
	}
}

func TestNumDivByZero(t *testing.T) {
	_, err := NumDiv(NumFromFloat(1), NumFromFloat(0))
	if err == nil {
		t.Error("expected divide by zero error")
	}
}

func TestNumSqrt(t *testing.T) {
	r, err := NumSqrt(NumFromFloat(144))
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(r.ToFloat64()-12) > 0.001 {
		t.Errorf("sqrt(144) = %f, want 12", r.ToFloat64())
	}
}

func TestNumSqrtNegative(t *testing.T) {
	_, err := NumSqrt(NumFromFloat(-4))
	if err == nil {
		t.Error("expected error for sqrt of negative")
	}
}

func TestNumSquare(t *testing.T) {
	r := NumSquare(NumFromFloat(7))
	if math.Abs(r.ToFloat64()-49) > 0.001 {
		t.Errorf("7^2 = %f, want 49", r.ToFloat64())
	}
}

func TestNumLog10(t *testing.T) {
	r, err := NumLog10(NumFromFloat(1000))
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(r.ToFloat64()-3) > 0.001 {
		t.Errorf("log10(1000) = %f, want 3", r.ToFloat64())
	}
}

func TestNumLog10Negative(t *testing.T) {
	_, err := NumLog10(NumFromFloat(-5))
	if err == nil {
		t.Error("expected error for log10 of negative")
	}
}

func TestSanitize(t *testing.T) {
	n := Number{Value: math.NaN(), Base: 0}
	if err := n.Sanitize(); err == nil {
		t.Error("expected NaN error")
	}
	n = Number{Value: math.Inf(1), Base: 0}
	if err := n.Sanitize(); err == nil {
		t.Error("expected Inf error")
	}
	n = Number{Value: 3.14, Base: 0}
	if err := n.Sanitize(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCmp(t *testing.T) {
	a := NumFromFloat(100)
	b := NumFromFloat(200)
	if a.Cmp(b) >= 0 {
		t.Error("100 should be < 200")
	}
	if b.Cmp(a) <= 0 {
		t.Error("200 should be > 100")
	}
	if a.Cmp(a) != 0 {
		t.Error("100 should equal 100")
	}
}

func TestDisplay(t *testing.T) {
	n := NumFromFloat(3140)
	d := n.Display()
	if d != "3140" {
		t.Errorf("Display(3140) = %q, want '3140'", d)
	}
}

func TestScientificNotation(t *testing.T) {
	// 3.14 * 10^100 should display in scientific notation
	n := Number{Value: 3.14, Base: 100}
	d := n.Display()
	if d != "3.14e100" {
		t.Errorf("Display = %q, want '3.14e100'", d)
	}
}

func TestNumFromFloat(t *testing.T) {
	n := NumFromFloat(0)
	if !n.IsZero() {
		t.Error("NumFromFloat(0) should be zero")
	}
	n = NumFromFloat(42)
	if math.Abs(n.ToFloat64()-42) > 0.001 {
		t.Errorf("NumFromFloat(42) = %f", n.ToFloat64())
	}
}
