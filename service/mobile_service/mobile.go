package mobile_service

import (
	"regexp"
	"strings"

	"github.com/nyaruka/phonenumbers"
	"github.com/smartblock/gta-api/pkg/e"
)

// ParseMobileNo func
func ParseMobileNo(mobileNo, countryCode string) (*phonenumbers.PhoneNumber, string) {
	var ok bool
	num, err := phonenumbers.Parse(mobileNo, countryCode)
	if err != nil {
		return nil, e.GetMsg(e.INVALID_MOBILE_NO)
	}

	ok = phonenumbers.IsPossibleNumber(num)
	if !ok {
		return nil, e.GetMsg(e.INVALID_MOBILE_NO)
	}

	ok = phonenumbers.IsValidNumber(num)
	if !ok {
		return nil, e.GetMsg(e.INVALID_MOBILE_NO)
	}

	return num, ""
}

// E164Format func
func E164Format(number *phonenumbers.PhoneNumber) string {
	return phonenumbers.Format(number, phonenumbers.E164)
}

// E164FormatWithouSymbol func
func E164FormatWithouSymbol(number *phonenumbers.PhoneNumber) string {
	num := E164Format(number)
	re := regexp.MustCompile("[0-9]+")
	s := re.FindAllString(num, -1) // get number only
	return strings.Join(s[:], "")
}

// NationalFormat func
func NationalFormat(number *phonenumbers.PhoneNumber) string {
	return phonenumbers.Format(number, phonenumbers.NATIONAL)
}

// NationalFormatWithouSymbol func
func NationalFormatWithouSymbol(number *phonenumbers.PhoneNumber) string {
	num := NationalFormat(number)
	re := regexp.MustCompile("[0-9]+")
	s := re.FindAllString(num, -1) // get number only
	return strings.Join(s[:], "")
}

// InternationalFormat func
func InternationalFormat(number *phonenumbers.PhoneNumber) string {
	return phonenumbers.Format(number, phonenumbers.INTERNATIONAL)
}

// RFC3966Format func
func RFC3966Format(number *phonenumbers.PhoneNumber) string {
	return phonenumbers.Format(number, phonenumbers.INTERNATIONAL)
}
