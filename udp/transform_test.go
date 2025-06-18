package udp

import (
	"slices"
	"testing"
)

func TestSplitLine(t *testing.T) {
	type testCase struct {
		Line  string
		Parts []string
	}
	cases := []testCase{
		{`pos_z v=1.222500 19523`, []string{"pos_z", "v=1.222500", "19523"}},
		{`fsensor error="value too long" 22575`, []string{"fsensor", `error="value too long"`, "22575"}},
		{`xbe_fan,fan=1 pwm=0i,rpm=0i 23427`, []string{"xbe_fan,fan=1", "pwm=0i,rpm=0i", "23427"}},
	}

	for _, tc := range cases {
		got := splitLine(tc.Line)

		if !slices.Equal(got, tc.Parts) {
			t.Errorf("splitLine(%q): got %v, want %v", tc.Line, got, tc.Parts)
		}
	}
}
