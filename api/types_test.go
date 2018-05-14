package api

import (
	"testing"
)

func TestControlInfo(t *testing.T) {
	ci := &ControlInfo{
		Power:    true,
		Mode:     "heat",
		Temp:     22.245,
		Humidity: 10,
		FanRate:  0,
		FanDir:   -1,
	}

	v := ci.Values()
	parsed := ParseControlInfo(v)

	if !parsed.Equal(ci) {
		t.Errorf("expected equal, parsed=%+v source=%+v", parsed, ci)
	}
}
