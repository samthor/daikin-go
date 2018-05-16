package api

import (
	"testing"
)

func TestConst(t *testing.T) {
	if FanRateAuto != -1 {
		t.Errorf("expected FanRateAuto=-1, was %v", FanRateAuto)
	}
	if FanRateQuiet != -2 {
		t.Errorf("expected FanRateQuiet=-2, was %v", FanRateQuiet)
	}
	if FanRateTwo != 2 {
		t.Errorf("expected FanRateTwo=2, was %v", FanRateTwo)
	}
	if FanDirNone != 1 {
		t.Errorf("expected FanDirNone=1, was %v", FanDirNone)
	}
}

func TestControlInfo(t *testing.T) {
	src := []*ControlInfo{
		&ControlInfo{
			Power:    true,
			Mode:     "heat",
			Temp:     22.5, // floating-point would be clamped
			Humidity: 10,
			FanRate:  FanRateAuto,
			FanDir:   FanDirNone,
		},
		&ControlInfo{
			Power:   true,
			Mode:    "auto",
			Temp:    20.0,
			FanRate: FanRateQuiet,
		},
		&ControlInfo{
			Power: true,
			Mode:  "cool",
			Temp:  18.0,
		},
		&ControlInfo{
			Power: false,
			Mode:  "auto",
			Temp:  18.0,
		},
	}

	for _, ci := range src {
		v := ci.Values()
		parsed := ParseControlInfo(v)
		if !parsed.Equal(ci) {
			t.Errorf("expected equal, parsed=%+v source=%+v values=%+v", parsed, ci, v)
		}
	}
}

func TestQuiet(t *testing.T) {
	v := (&ControlInfo{Power: true, FanRate: FanRateQuiet}).Values()
	if v.Get("f_rate") != "B" {
		t.Errorf("expected quiet to be \"B\", was: %v", v.Get("f_rate"))
	}
}
