package aut_test

import (
	"strings"
	"testing"

	"go.nickng.io/sesstype/local"
	"go.nickng.io/sesstypeconv/internal/aut"
)

func TestBranch(t *testing.T) {
	l, err := local.Parse(strings.NewReader(`
	A&{
		?l().end,
		?(int).end,
		?l2(B&{?L(T).end}).end
	}`))
	if err != nil {
		t.Fatal(err)
	}
	a := aut.FromLocal(l)
	if want, got := 4, a.NumStates; want != got {
		t.Log(a.String())
		t.Errorf("expected %d states but got %d", want, got)
	}
	if want, got := 3, a.NumTransitions; want != got {
		t.Log(a.String())
		t.Errorf("expected %d transitions but got %d", want, got)
	}
	r, err := aut.ToLocal(a)
	if !local.Equal(l, r) {
		t.Logf("\n%s\n%s\n", l.String(), r.String())
		t.Errorf("local type ↔ aut round-trip conversion should be equal")
	}
}

func TestSelect(t *testing.T) {
}

func TestRecur(t *testing.T) {
	l, err := local.Parse(strings.NewReader(`
	*L0.A?().*L1.A&{
		?l().L1,
		?(int).L0,
		?l2(B&{?L(T).end}).end
	}`))
	if err != nil {
		t.Fatal(err)
	}
	a := aut.FromLocal(l)
	if want, got := 3, a.NumStates; want != got {
		t.Log(a.String())
		t.Errorf("expected %d states but got %d", want, got)
	}
	r, err := aut.ToLocal(a)
	if !local.Equal(l, r) {
		t.Logf("\n%s\n%s\n", l.String(), r.String())
		t.Errorf("local type ↔ aut round-trip conversion should be equal")
	}
}

func TestRecurSimplified(t *testing.T) {
	l, err := local.Parse(strings.NewReader(`
	*T.*x.A&{
		?l().x,
		?(int).T,
		?l2(B&{?L(T).end}).end
	}`))
	if err != nil {
		t.Fatal(err)
	}
	a := aut.FromLocal(l)
	if want, got := 2, a.NumStates; want != got {
		t.Log(a.String())
		t.Errorf("expected %d states but got %d", want, got)
	}
	if want, got := 3, a.NumTransitions; want != got {
		t.Log(a.String())
		t.Errorf("expected %d transitions but got %d", want, got)
	}
	r, err := aut.ToLocal(a)
	simplified, err := local.Parse(strings.NewReader(`*L0.A&{?l().L0,?(int).L0, ?l2(B&{?L(T).end}).end}`))
	if err != nil {
		t.Fatal(err)
	}
	if !local.Equal(simplified, r) {
		t.Logf("\n%s\n%s\n", simplified, r)
		t.Errorf("local type ↔ aut round-trip conversion should be equal")
	}
}

func TestEnd(t *testing.T) {
	l, err := local.Parse(strings.NewReader(`end`))
	if err != nil {
		t.Fatal(err)
	}
	a := aut.FromLocal(l)
	if want, got := 1, a.NumStates; want != got {
		t.Log(a.String())
		t.Errorf("expected %d states but got %d", want, got)
	}
	r, err := aut.ToLocal(a)
	if !local.Equal(l, r) {
		t.Logf("\n%s\n%s\n", l.String(), r.String())
		t.Errorf("local type ↔ aut round-trip conversion should be equal")
	}
}
