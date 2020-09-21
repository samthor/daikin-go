package api

import (
	"fmt"
	"net/url"
	"strconv"
)

const (
	tempMin = 19.0 // 10.0
	tempMax = 25.0 // 41.0
)

// FanRate indicates the rate of the fan.
type FanRate int

const (
	FanRateUnset FanRate = iota
	FanRateOne
	FanRateTwo
	FanRateThree
	FanRateFour
	FanRateFive

	FanRateQuiet FanRate = -2
	FanRateAuto  FanRate = -1
)

// FanDir indicates what direction the fan should run.
type FanDir int

const (
	FanDirUnset FanDir = iota
	FanDirNone
	FanDirVertical
	FanDirHorizontal
	FanDirBoth
)

// ParseControlInfo parses a ControlInfo from a url.Values.
func ParseControlInfo(v url.Values) (out ControlInfo) {
	out.Power = v.Get("pow") == "1"

	out.Mode = "auto"
	mode, _ := strconv.Atoi(v.Get("mode"))
	switch mode {
	case 2:
		out.Mode = "dehum"
	case 3:
		out.Mode = "cool"
	case 4:
		out.Mode = "heat"
	case 6:
		out.Mode = "fan"
	}

	out.Temp, _ = strconv.ParseFloat(v.Get("stemp"), 64)
	out.Humidity = parseHumidity(v.Get("shum"))
	out.FanRate = parseFanRate(v.Get("f_rate"))
	out.FanDir = parseFanDir(v.Get("f_dir"))

	out.OtherModes = map[string]ControlInfoMode{}

	updateMode := func(index int, name string) {
		var info ControlInfoMode
		get := func(name string) string {
			return v.Get(fmt.Sprintf("%s%d", name, index))
		}

		info.Temp, _ = strconv.ParseFloat(get("dt"), 64)
		info.Humidity = parseHumidity(get("dh"))
		info.FanRate = parseFanRate(get("dfr"))
		info.FanDir = parseFanDir(get("dfd"))

		out.OtherModes[name] = info
	}

	updateMode(1, "auto")
	updateMode(2, "dehum")
	updateMode(3, "cool")
	updateMode(4, "heat")
	updateMode(6, "fan")

	return
}

func parseHumidity(raw string) int {
	out, err := strconv.Atoi(raw)
	if err != nil || out < 0 || out > 100 {
		return -1
	}
	return out
}

func parseFanRate(raw string) FanRate {
	if frate, err := strconv.Atoi(raw); err == nil {
		if frate >= 3 && frate <= 7 {
			return FanRate(frate - 2) // daikin uses 3-7 for rates 1-5
		}
	} else if raw == "A" {
		return FanRateAuto
	} else if raw == "B" {
		return FanRateQuiet
	}
	return FanRateUnset
}

func parseFanDir(raw string) FanDir {
	if fdir, err := strconv.Atoi(raw); err == nil {
		return FanDir(fdir) + 1 // our values are 1-4, device 0-3
	}
	return FanDirUnset
}

// SensorInfo contains sensor information from a Daikin AC.
type SensorInfo struct {
	Temp  float64  // inside
	World *float64 // outside
}

// ParseSensorInfo parses a SensorInfo from a url.Values.
func ParseSensorInfo(v url.Values) (out SensorInfo) {
	out.Temp, _ = strconv.ParseFloat(v.Get("htemp"), 64)
	if otemp, err := strconv.ParseFloat(v.Get("otemp"), 64); err == nil {
		out.World = &otemp
	}
	return
}

// ControlInfo specifies how to interact with a Daikin AC.
type ControlInfo struct {
	Power bool
	Mode  string // one of "auto", "dehum", "cool", "heat", or "fan"

	ControlInfoMode // embedded/target

	OtherModes map[string]ControlInfoMode // info from get
}

// ControlInfoMode is returned from the AC and represents the state of other modes.
type ControlInfoMode struct {
	Temp     float64 // -ve for "M"
	Humidity int     // 0-50, -ve for "AUTO"
	FanRate  FanRate
	FanDir   FanDir
}

// Equal determines whether this ControlInfo is equal to another ControlInfo.
// Nil instances are equal, as well as 'powered off' instances.
func (ci *ControlInfo) Equal(other *ControlInfo) bool {
	if ci == nil {
		return other == nil
	} else if !ci.Power {
		return !other.Power
	}

	return ci.Mode == other.Mode &&
		ci.clampTemp() == other.clampTemp() &&
		ci.Humidity == other.Humidity &&
		ci.FanRate == other.FanRate &&
		ci.FanDir == other.FanDir
}

func (ci *ControlInfo) clampTemp() float64 {
	t := ci.Temp
	if t < 0.0 {
		return -1.0 // "M" mode
	} else if t == 0.0 {
		return 0.0 // default/nothing
	} else if t <= tempMin {
		return tempMin
	} else if t >= tempMax {
		return tempMax
	}
	// clamp to 0.5
	value := int(t * 2)
	return float64(value) / 2.0
}

// Values converts this ControlInfo into url.Values to send to the Daikin.
func (ci *ControlInfo) Values() url.Values {
	v := url.Values{}

	// mode: 0-7
	mode := 0
	switch ci.Mode {
	case "dehum":
		mode = 2
	case "cool":
		mode = 3
	case "heat":
		mode = 4
	case "fan":
		mode = 6
	}
	v.Set("mode", fmt.Sprintf("%d", mode))

	// stemp: 10.0-41.0
	temp := ci.clampTemp()
	if temp < 0.0 {
		v.Set("stemp", "M") // TODO: still not sure what this does
	} else {
		v.Set("stemp", strconv.FormatFloat(temp, 'f', 1, 64))
	}

	// shum: 0-50 or "AUTO"
	if ci.Humidity < 0 {
		v.Set("shum", "AUTO")
	} else {
		hum := ci.Humidity
		if hum > 50 {
			hum = 50
		}
		v.Set("shum", fmt.Sprintf("%d", hum))
	}

	// power: 0-1
	if !ci.Power {
		v.Set("pow", "0")
		return v // rest is optional
	}
	v.Set("pow", "1")

	// f_rate (optional): 1-5, "A" (auto) or "B" (silent)
	if ci.FanRate == FanRateAuto {
		v.Set("f_rate", "A")
	} else if ci.FanRate == FanRateQuiet {
		v.Set("f_rate", "B")
	} else if ci.FanRate >= 1 {
		rate := ci.FanRate
		if rate > 5 {
			rate = 5
		}
		v.Set("f_rate", fmt.Sprintf("%d", rate+2)) // rates are 3-7
	}

	// f_dir (optional): 0-3
	if ci.FanDir >= 1 && ci.FanDir <= 4 {
		v.Set("f_dir", fmt.Sprintf("%d", ci.FanDir-1))
	}

	return v
}
