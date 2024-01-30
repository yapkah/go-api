package base

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// GetTimeZone func
func GetTimeZone() *time.Location {
	loc, _ := time.LoadLocation("Asia/Kuala_Lumpur")
	return loc
}

//GetCurrentTime func
func GetCurrentTime(format string) string {
	//init the loc
	loc := GetTimeZone()

	//set timezone
	now := time.Now().In(loc).Format(format)

	return now
}

//GetCurrentDateTimeT func
func GetCurrentDateTimeT() time.Time {
	//init the loc
	loc := GetTimeZone()

	//set timezone
	now := time.Now().In(loc)

	return now
}

// TimeFormat func
func TimeFormat(time time.Time, format string) string {
	//init the loc
	loc := GetTimeZone()

	//set timezone
	date := time.In(loc).Format(format)

	return date
}

//ccccc
const (
	yyyy = "2006"
	yy   = "06"
	mmmm = "January"
	mmm  = "Jan"
	mm   = "01"
	dddd = "Monday"
	ddd  = "Mon"
	dd   = "02"

	HHT = "03"
	HH  = "15"
	MM  = "04"
	SS  = "05"
	ss  = "05"
	tt  = "PM"
	Z   = "MST"
	ZZZ = "MST"

	o = "Z07:00"
)

//GetCurrentTimeV2 func
func GetCurrentTimeV2(format string) (string, error) {
	tf := ConvertFormat(format)

	//init the loc
	loc := GetTimeZone()

	//set timezone
	now := time.Now().In(loc).Format(tf)

	// err = &e.CustomError{HTTPCode: http.StatusBadRequest, Code: e.INVALID_VERSION_NUMBER_FORMAT}

	return now, nil
}

//ConvertFormat func
func ConvertFormat(format string) string {
	var newFormat = format
	if strings.Contains(newFormat, "YYYY") {
		newFormat = strings.Replace(newFormat, "YYYY", yyyy, -1)
	} else if strings.Contains(newFormat, "yyyy") {
		newFormat = strings.Replace(newFormat, "yyyy", yyyy, -1)
	} else if strings.Contains(newFormat, "YY") {
		newFormat = strings.Replace(newFormat, "YY", yy, -1)
	} else if strings.Contains(newFormat, "yy") {
		newFormat = strings.Replace(newFormat, "yy", yy, -1)
	}

	if strings.Contains(newFormat, "MMMM") {
		newFormat = strings.Replace(newFormat, "MMMM", mmmm, -1)
	} else if strings.Contains(newFormat, "mmmm") {
		newFormat = strings.Replace(newFormat, "mmmm", mmmm, -1)
	} else if strings.Contains(newFormat, "MMM") {
		newFormat = strings.Replace(newFormat, "MMM", mmm, -1)
	} else if strings.Contains(newFormat, "mmm") {
		newFormat = strings.Replace(newFormat, "mmm", mmm, -1)
	} else if strings.Contains(newFormat, "mm") {
		newFormat = strings.Replace(newFormat, "mm", mm, -1)
	}

	if strings.Contains(newFormat, "dddd") {
		newFormat = strings.Replace(newFormat, "dddd", dddd, -1)
	} else if strings.Contains(newFormat, "ddd") {
		newFormat = strings.Replace(newFormat, "ddd", ddd, -1)
	} else if strings.Contains(newFormat, "dd") {
		newFormat = strings.Replace(newFormat, "dd", dd, -1)
	}

	if strings.Contains(newFormat, "tt") {
		if strings.Contains(newFormat, "HH") {
			newFormat = strings.Replace(newFormat, "HH", HHT, -1)
		} else if strings.Contains(newFormat, "hh") {
			newFormat = strings.Replace(newFormat, "hh", HHT, -1)
		}
		newFormat = strings.Replace(newFormat, "tt", tt, -1)
	} else {
		if strings.Contains(newFormat, "HH") {
			newFormat = strings.Replace(newFormat, "HH", HH, -1)
		} else if strings.Contains(newFormat, "hh") {
			newFormat = strings.Replace(newFormat, "hh", HH, -1)
		}
		newFormat = strings.Replace(newFormat, "tt", "", -1)
	}

	if strings.Contains(newFormat, "MM") {
		newFormat = strings.Replace(newFormat, "MM", MM, -1)
	}

	if strings.Contains(newFormat, "SS") {
		newFormat = strings.Replace(newFormat, "SS", SS, -1)
	} else if strings.Contains(newFormat, "ss") {
		newFormat = strings.Replace(newFormat, "ss", SS, -1)
	}

	if strings.Contains(newFormat, "ZZZ") {
		newFormat = strings.Replace(newFormat, "ZZZ", ZZZ, -1)
	} else if strings.Contains(newFormat, "zzz") {
		newFormat = strings.Replace(newFormat, "zzz", ZZZ, -1)
	} else if strings.Contains(newFormat, "Z") {
		newFormat = strings.Replace(newFormat, "Z", Z, -1)
	} else if strings.Contains(newFormat, "z") {
		newFormat = strings.Replace(newFormat, "z", Z, -1)
	}

	if strings.Contains(newFormat, "tt") {
		newFormat = strings.Replace(newFormat, "tt", tt, -1)
	}
	if strings.Contains(newFormat, "o") {
		newFormat = strings.Replace(newFormat, "o", o, -1)
	}
	return newFormat
}

// StrToUnixTime func
func StrToUnixTime(str string) (int64, error) {
	layout := "2006-01-02 15:04:05"
	loc := GetTimeZone()
	t, err := time.ParseInLocation(layout, str, loc)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// StrToDateTime func
func StrToDateTime(str, format string) (time.Time, error) {
	loc := GetTimeZone()
	t, err := time.ParseInLocation(format, str, loc)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

type HrMinSec struct {
	Hours   float64
	Minutes float64
	Seconds float64
}

// ConvertSecToHrMinSec func
func ConvertSecToHrMinSec(duration time.Duration) HrMinSec {

	hour := math.Floor(duration.Hours())
	hourInMinutes := hour * 60
	hourInSeconds := hour * 60 * 60
	minutes := duration.Minutes() - hourInMinutes
	minutes = math.Floor(minutes)
	MinutesInSeconds := minutes * 60

	seconds := math.Floor(duration.Seconds() - MinutesInSeconds - hourInSeconds)

	// fmt.Println("hour:", hour)
	// fmt.Println("minutes:", minutes)
	// fmt.Println("seconds:", seconds)
	arrDataReturn := HrMinSec{
		Hours:   hour,
		Minutes: minutes,
		Seconds: seconds,
	}

	return arrDataReturn
}

// func AddDurationInString
func AddDurationInString(timeStart time.Time, extendDuration string) time.Time {

	extendedDT := timeStart
	if strings.Contains(extendDuration, "minute") || strings.Contains(extendDuration, "minutes") { // perform add number of minute / minutes
		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
		submatchall := re.FindAllString(extendDuration, -1)
		duration := 0
		durationSetting, err := strconv.Atoi(submatchall[0])
		if err == nil {
			duration = durationSetting
		}
		extendedDT = timeStart.Add(time.Minute * time.Duration(duration))
	} else if strings.Contains(extendDuration, "hour") || strings.Contains(extendDuration, "hours") { // perform add number of hour / hours
		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
		submatchall := re.FindAllString(extendDuration, -1)
		duration := 0
		durationSetting, err := strconv.Atoi(submatchall[0])
		if err == nil {
			duration = durationSetting
		}
		extendedDT = timeStart.Add(time.Hour * time.Duration(duration))
	} else if strings.Contains(extendDuration, "day") || strings.Contains(extendDuration, "days") { // perform add number of day / days
		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
		submatchall := re.FindAllString(extendDuration, -1)
		duration := 0
		durationSetting, err := strconv.Atoi(submatchall[0])
		if err == nil {
			duration = durationSetting
		}
		extendedDT = timeStart.AddDate(0, 0, duration)
	}

	return extendedDT
}

// ValidateDateTimeFormat func
func ValidateDateTimeFormat(dateTime string, format string) (string, bool) {
	convertedDateTime, err := StrToDateTime(dateTime, format)

	if err != nil {
		return "", false
	}

	dateTimeTrimmedStr := convertedDateTime.Format(format)

	return dateTimeTrimmedStr, true
}
