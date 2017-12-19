package sesstypeconv_test

import (
	"strings"
	"testing"

	"go.nickng.io/sesstype/local"
	"go.nickng.io/sesstypeconv"
)

func TestToAut(t *testing.T) {
	l, err := local.Parse(strings.NewReader(`*x.A&{?().end, ?local(int).x}`))
	if err != nil {
		t.Errorf("Cannot parse local type: %v", err)
	}
	a := sesstypeconv.ToAut(l)
	t.Log(a.String())
	out1 := `des (0, 2, 3)
(0, A ? (), 1)
(0, A ? local(int), 2)
`
	out2 := `des (0, 2, 3)
(0, A ? local(int), 2)
(0, A ? (), 1)
`
	if a.String() != out1 && a.String() != out2 {
		t.Logf("expecting:\n%s\nor:\n%s\nbut got %s", out1, out2, a.String())
	}
}
