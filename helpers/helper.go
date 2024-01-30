package helpers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/dustin/go-humanize"
	"github.com/smartblock/gta-api/models"
	"github.com/smartblock/gta-api/pkg/float"
	"github.com/smartblock/gta-api/pkg/setting"

	//"golang.org/x/text/language"
	//"golang.org/x/text/message"
	"hash/fnv"
	"io/ioutil"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"unsafe"
)

/**

This is a file to place helper functions unrelated to database models
Please do not import models in here so it can be imported to models file

*/

type SMarketPriceStruc struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type FMarketPriceStruc struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
}

func Contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func IsNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func NumberFormat2Dec(s string) float64 {
	val, _ := strconv.ParseFloat(s, 64)
	return math.Floor(val*100) / 100
}

func NumberFormatInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

func NumberFormatInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

func BytesToString(b []byte) string {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := reflect.StringHeader{Data: bh.Data, Len: bh.Len}
	return *(*string)(unsafe.Pointer(&sh))
}

// FNV32a hashes using fnv32a algorithm
func FNV32a(text string) uint32 {
	algorithm := fnv.New32a()
	algorithm.Write([]byte(text))
	return algorithm.Sum32()
}

// get market price from binance
func GetBinanceMarketPrice(WalletType string) (float64, string) {
	var s_market_price SMarketPriceStruc

	if WalletType == "USDT" {
		s_market_price.Price = "1"
	} else {
		//detail, _ := models.GetSysSettingByID("live_price", "binance_live_price")
		resp, _ := http.Get("https://api.binance.com/api/v3/ticker/price?symbol=" + WalletType + "USDT")
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		_ = json.Unmarshal([]byte(string(body)), &s_market_price)
	}

	f_market_price, _ := strconv.ParseFloat(s_market_price.Price, 64)

	return f_market_price, s_market_price.Price
}

// slice decimal without rounding off, d must always equal or greater then zero
func NumberFormat(value float64, d int) (f float64) {
	m := math.Mod(value*math.Pow10(d), (1/math.Pow10(d))*math.Pow10(d))

	f = value - (m / math.Pow10(d))
	return f
}

/*helpers.BalanceFormat(helpers.NumberFormat2Dec("123123.800"), 8)*/
func BalanceFormat(value float64, d int) (s string) {
	//value = float64(int(value * math.Pow10(d) + (1 * math.Pow10(d))) - int(1 * math.Pow10(d))) / math.Pow10(d)
	//use big.Float to avoid float precision calculation
	const prec = 200
	str := fmt.Sprintf("%f", value) // s == "123.456000"
	a, _ := new(big.Float).SetPrec(prec).SetString(str)

	var format string
	format = "#,###."
	for i := 0; i < d; i++ {
		format = format + "#"
	}
	f, _ := strconv.ParseFloat(a.String(), 64)
	s = humanize.FormatFloat(format, f)

	return s
}

func TrailZeroFloat(value float64, d int) (s float64) {
	value = float64(int(value * math.Pow10(d)))

	return value
}

func DetrailZeroFloat(value float64, d int) (s float64) {
	value = value / math.Pow10(d)

	return value
}

func Textify(value float64) (s string) {
	var dom float64
	var unit string

	if value < 1000 {
		dom = value
		unit = ""
	} else if value >= 1000 && value < 1000000 {
		dom = value / 1000
		unit = "K"
	} else {
		dom = value / 1000000
		unit = "M"
	}

	s = fmt.Sprintf("%g", dom) + unit
	return s
}

func GenCheckSum(memId string, walletId string, docNo string, transactionType string, gameId string, date string) (string, error) {
	hashString := memId + walletId + docNo + transactionType + gameId + date + "Slot777SM"

	hasher := md5.New()
	hasher.Write([]byte(hashString))
	hash := hex.EncodeToString(hasher.Sum(nil))

	return hash, nil
}

func GetCurrentTime() string {
	//init the loc
	loc, _ := time.LoadLocation("Asia/Kuala_Lumpur")

	//set timezone
	now := time.Now().In(loc).Format("2006-01-02 15:04:05.000000")

	return now
}

func ConvertTimeToUnix(value time.Time) string {
	t := value.Unix()
	return strconv.FormatInt(int64(t), 10)
}

func Translate(word string, langCode string) string {
	// check language
	ok := models.ExistLangague(langCode)
	if !ok {
		langCode = "en"
	}

	word = strings.Replace(strings.ToLower(word), " ", "_", -1)
	dbWord, _ := models.GetTranslationByName(langCode, "label", "label."+word)
	translatedWord := word

	if dbWord == nil {
		AddTranslationV2("label." + word)
	}

	if dbWord != nil {
		translatedWord = dbWord.Value
	}
	return translatedWord
}

// TranslateV2 version 2: support replace value
func TranslateV2(word string, langCode string, params map[string]string) string {
	// check language
	ok := models.ExistLangague(langCode)
	if !ok {
		langCode = setting.Cfg.Section("app").Key("DefaultLangCode").String()
	}

	word = strings.Replace(strings.ToLower(word), " ", "_", -1)
	dbWord, _ := models.GetTranslationByName(langCode, "label", "label."+word)
	translatedWord := word

	if dbWord == nil {
		AddTranslationV2("label." + word)
	}

	if dbWord != nil {
		translatedWord = dbWord.Value
	}
	if len(params) > 0 {
		for k1, v1 := range params {
			strToFind := ":" + k1
			if strings.Contains(k1, "_translate") {
				v1 = strings.Replace(strings.ToLower(v1), " ", "_", -1)
				dbWordV, _ := models.GetTranslationByName(langCode, "label", "label."+v1)
				if dbWordV != nil {
					v1 = dbWordV.Value
				}
				strToFind = strings.Replace(strToFind, "_translate", "", -1)
				translatedWord = strings.Replace(translatedWord, strToFind, v1, -1)
			} else {
				translatedWord = strings.Replace(translatedWord, strToFind, v1, -1)
			}
		}
	}
	return translatedWord
}

// ValueToInt convert value to Int
func ValueToInt(value string) (int, error) {
	data, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return data, nil
}

// ValueToBool convert value to boolean
func ValueToBool(value string) (bool, error) {
	integer, err := ValueToInt(value)
	if err != nil {
		return false, err
	}

	if integer == 0 {
		return false, nil
	}

	return true, nil
}

// ValueToFloat convert value to float
func ValueToFloat(value string) (float64, error) {
	data, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	return data, nil
}

// ValueToDuration convert value to duration
func ValueToDuration(value string) (time.Duration, error) {
	data, err := ValueToInt(value)
	if err != nil {
		return 0, err
	}
	return time.Duration(data), nil
}

// StringInSlice verify if string in slice
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// StringToSliceInt convert string into slice of int. example 9,10 into []int{9,10}
func StringToSliceInt(stringList string, separator string) []int {
	if stringList != "" {
		sliceOfString := strings.Split(stringList, separator)
		b := make([]int, len(sliceOfString))
		for i, v := range sliceOfString {
			integer, err := strconv.Atoi(v)
			if err != nil {
				return nil
			}
			b[i] = integer
		}
		return b
	}
	return nil
}

func TransRemark(word string, langCode string) string {
	// check language
	ok := models.ExistLangague(langCode)
	if !ok {
		langCode = "en"
	}

	newWord := word

	re := regexp.MustCompile(`\#\*(.*?)\*\#`)

	submatchall := re.FindAllString(word, -1)
	for _, element := range submatchall {

		ori := element

		element = strings.Trim(element, "#*")
		element = strings.Trim(element, "*#")
		word = strings.Replace(strings.ToLower(element), " ", "_", -1)
		dbWord, _ := models.GetTranslationByName(langCode, "label", "label."+word)
		translatedWord := word

		// temp
		if dbWord == nil {
			AddTranslationV2("label." + word)
		}

		if dbWord != nil {
			translatedWord = dbWord.Value
		}

		newWord = strings.Replace(newWord, ori, translatedWord, 1)

	}

	return newWord
}

// ReverseSlice reverse a slice order
// func ReverseSlice(list interface{}) interface{} {
// 	listVal := reflect.ValueOf(list)

// 	// for i, j := 0, len(arrHistoryList)-1; i < j; i, j = i+1, j-1 {
// 	// 	arrHistoryList[i], arrHistoryList[j] = arrHistoryList[j], arrHistoryList[i]
// 	// }

// 	for i, j := 0, listVal.Len()-1; i < j; i, j = i+1, j-1 {
// 		listVal.Index(i), listVal.Index(j) = listVal.Index(j), listVal.Index(i)
// 	}

// 	return listVal
// }

// TransDiamondName translate diamond name dynamically
func TransDiamondName(diamondName, langCode string) string {
	regex := regexp.MustCompile("[0-9]+")
	matches := regex.FindAllString(diamondName, -1)

	if len(matches) > 0 {
		labelName := strings.Replace(strings.Replace(strings.ToLower(diamondName), " ", "_", -1), matches[0], ":num", -1)

		arrTransValue := make(map[string]string)
		arrTransValue["num"] = matches[0]

		diamondName = TranslateV2(labelName, langCode, arrTransValue)
	}

	return diamondName
}

// Paginate function
func Paginate(pageNum int, pageSize int, sliceLength int) (int, int) {
	start := pageNum * pageSize

	if start > sliceLength {
		start = sliceLength
	}

	end := start + pageSize
	if end > sliceLength {
		end = sliceLength
	}

	return start, end
}

// IntInSlice verify if int in slice
func IntInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Float64InSlice verify if int in slice
func Float64InSlice(a float64, list []float64) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// func NumberFormatPhp. this is work like php number_format
func NumberFormatPhp(number float64, decimals uint, decPoint, thousandsSep string) string {
	neg := false
	if number < 0 {
		number = -number
		neg = true
	}
	dec := int(decimals)
	// Will round off
	decimalFormat := "%." + strconv.Itoa(dec) + "f"
	// fmt.Println("decimalFormat:", decimalFormat)
	str := fmt.Sprintf(decimalFormat, number)
	// str := strconv.FormatFloat(number, 'f', dec, 64)
	// fmt.Println("str:", str)
	prefix, suffix := "", ""
	if dec > 0 {
		prefix = str[:len(str)-(dec+1)]
		suffix = str[len(str)-dec:]
	} else {
		prefix = str
	}
	sep := []byte(thousandsSep)
	n, l1, l2 := 0, len(prefix), len(sep)
	// thousands sep num
	c := (l1 - 1) / 3
	tmp := make([]byte, l2*c+l1)
	pos := len(tmp) - 1
	for i := l1 - 1; i >= 0; i, n, pos = i-1, n+1, pos-1 {
		if l2 > 0 && n > 0 && n%3 == 0 {
			for j := range sep {
				tmp[pos] = sep[l2-j-1]
				pos--
			}
		}
		tmp[pos] = prefix[i]
	}
	s := string(tmp)
	if dec > 0 {
		s += decPoint + suffix
	}
	if neg {
		s = "-" + s
	}

	return s
}

// func CutOffDecimalBK
func CutOffDecimalBK(number float64, decimals uint, decPoint, thousandsSep string) string {
	neg := false
	if number < 0 {
		number = -number
		neg = true
	}
	dec := int(decimals) + 2
	// Will round off
	number = float.RoundDown(number, dec) // round down decimal
	fmt.Println("number:", number)
	str := fmt.Sprintf("%."+strconv.Itoa(dec)+"F", number)
	fmt.Println("str:", str)
	prefix, suffix := "", ""
	if dec > 0 {
		prefix = str[:len(str)-(dec+1)]
		suffix = str[len(str)-dec:]
	} else {
		prefix = str
	}
	sep := []byte(thousandsSep)
	n, l1, l2 := 0, len(prefix), len(sep)
	// thousands sep num
	c := (l1 - 1) / 3
	tmp := make([]byte, l2*c+l1)
	pos := len(tmp) - 1
	for i := l1 - 1; i >= 0; i, n, pos = i-1, n+1, pos-1 {
		if l2 > 0 && n > 0 && n%3 == 0 {
			for j := range sep {
				tmp[pos] = sep[l2-j-1]
				pos--
			}
		}
		tmp[pos] = prefix[i]
	}
	s := string(tmp)
	if dec > 0 {
		s += decPoint + suffix
	}
	if neg {
		s = "-" + s
	}

	return s
}

func CutOffDecimalBK2(number float64, decimals uint, decPoint, thousandsSep string) string {

	numCutChar := 2
	dec := int(decimals) + numCutChar
	number = float.RoundDown(number, dec) // round down decimal
	// fmt.Println("number:", number)
	convertedDecimal := NumberFormatPhp(number, decimals+uint(numCutChar), decPoint, thousandsSep)
	// fmt.Println("convertedDecimal:", convertedDecimal)
	convertedDecimalWordCount := utf8.RuneCountInString(convertedDecimal)
	// fmt.Println("convertedDecimalWordCount:", convertedDecimalWordCount)
	if decimals == 0 {
		numCutChar = numCutChar + 1 // need to add 1 bcz of decimal
	}
	convertedDecimal = Substr(convertedDecimal, 0, convertedDecimalWordCount-numCutChar)

	return convertedDecimal
}

func CutOffDecimal(number float64, decimals uint, decPoint, thousandsSep string) string {
	neg := false
	if number < 0 {
		number = -number
		neg = true
	}
	dec := int(decimals)
	str := strconv.FormatFloat(float64(number), 'f', -1, 64)
	arrS := strings.Split(str, ".")
	convertedDecimal := "0" // default value is 0

	if decimals > 0 {
		if len(arrS) == 2 {
			s2 := ProcessNumberToDesireString(arrS[1], dec)
			str = arrS[0] + "." + s2
			convertedDecimal = ProcessNumberToDesirePart2String(neg, str, dec, decPoint, thousandsSep)
		} else if len(arrS) == 1 {
			defPart2String := ""
			s2 := ProcessNumberToDesireString(defPart2String, dec)
			if s2 != "" {
				str = arrS[0] + "." + s2
			}
			convertedDecimal = ProcessNumberToDesirePart2String(neg, str, dec, decPoint, thousandsSep)
		} else {
			defPart1String := "0"
			defPart2String := ""
			s2 := ProcessNumberToDesireString(defPart2String, dec)
			str = defPart1String
			if s2 != "" {
				str = defPart1String + "." + s2
			}
			convertedDecimal = ProcessNumberToDesirePart2String(neg, str, dec, decPoint, thousandsSep)
		}
	} else {
		convertedDecimal = ProcessNumberToDesirePart2String(neg, arrS[0], dec, decPoint, thousandsSep)
	}

	return convertedDecimal
}

func CutOffDecimalv2(number float64, decimals uint, decPoint, thousandsSep string, flexZeroDecPointStatus bool) string {
	neg := false
	if number < 0 {
		number = -number
		neg = true
	}
	dec := int(decimals)
	str := strconv.FormatFloat(float64(number), 'f', -1, 64)
	arrS := strings.Split(str, ".")
	convertedDecimal := "0" // default value is 0

	if decimals > 0 {
		if len(arrS) == 2 {
			s2 := arrS[1]
			if !flexZeroDecPointStatus {
				s2 = ProcessNumberToDesireString(arrS[1], dec)
			} else {
				if dec < len(s2) {
					s2 = Substr(s2, 0, dec)
				} else if dec > len(s2) {
					dec = len(s2)
				}
			}
			str = arrS[0] + "." + s2
			convertedDecimal = ProcessNumberToDesirePart2String(neg, str, dec, decPoint, thousandsSep)
		} else if len(arrS) == 1 {
			defPart2String := ""
			s2 := ""
			if !flexZeroDecPointStatus {
				s2 = ProcessNumberToDesireString(defPart2String, dec)
			} else {
				dec = 0
			}
			if s2 != "" {
				str = arrS[0] + "." + s2
			}
			convertedDecimal = ProcessNumberToDesirePart2String(neg, str, dec, decPoint, thousandsSep)
		} else {
			defPart1String := "0"
			defPart2String := ""
			s2 := ""
			if !flexZeroDecPointStatus {
				s2 = ProcessNumberToDesireString(defPart2String, dec)
			} else {
				dec = 0
			}
			str = defPart1String
			if s2 != "" {
				str = defPart1String + "." + s2
			}
			convertedDecimal = ProcessNumberToDesirePart2String(neg, str, dec, decPoint, thousandsSep)
		}
	} else {
		convertedDecimal = ProcessNumberToDesirePart2String(neg, arrS[0], dec, decPoint, thousandsSep)
	}

	return convertedDecimal
}

// Substr. this is work like php substr()
func Substr(str string, start uint, length int) string {
	if start < 0 || length < -1 {
		return str
	}
	switch {
	case length == -1:
		return str[start:]
	case length == 0:
		return ""
	}
	end := int(start) + length
	if end > len(str) {
		end = len(str)
	}
	return str[start:end]
}

func Explode(str, delimiter string) []string {
	return strings.Split(strings.ReplaceAll(str, " ", ""), delimiter)
}

func CutOffStringsDecimal(number string, decimals uint, decPoint byte) string {

	chck := strings.Contains(number, string(decPoint))

	length := len(number)

	convertedDecimal := number

	if chck {
		if length > int(decimals) {
			char := number[:strings.IndexByte(number, decPoint)]

			charLength := len(char)

			if decimals > 0 {
				charLength = charLength + 1 + int(decimals)

				convertedDecimal = Substr(number, 0, charLength)
			}
		}
	}

	return convertedDecimal
}

// IsArray func to check if interface is array
func IsArray(m map[string]interface{}) bool {
	for _, value := range m {
		// fmt.Println("k:", k, "value:", value)
		rt := reflect.TypeOf(value)

		switch rt.Kind() {
		case reflect.Slice:
			// fmt.Println(k, "is a slice with element type", rt.Elem())
			return true
		case reflect.Array:
			// fmt.Println(k, "is an array with element type", rt.Elem())
			return true
		case reflect.Map:
			// fmt.Println(k, "is a map with element type", rt.Elem())
			return true
		default:
			// fmt.Println(k, "is something else entirely")
			return false
		}
	}

	return false
}

// IsMultipleOf func
func IsMultipleOf(value, multipleOf float64) bool {
	dividedValue := float.Div(value, multipleOf)

	if dividedValue == float64(int(dividedValue)) {
		return true
	} else {
		return false
	}
}

func ProcessNumberToDesireString(number string, dec int) string {
	s2 := Substr(number, 0, dec)
	bal := dec - len(s2)
	if bal > 0 {
		for i := 0; i < bal; i++ {
			s2 = s2 + "0"
		}
	}
	return s2
}

func ProcessNumberToDesirePart2String(neg bool, str string, dec int, decPoint string, thousandsSep string) string {
	prefix, suffix := "", ""
	if dec > 0 {
		prefix = str[:len(str)-(dec+1)]
		suffix = str[len(str)-dec:]
	} else {
		prefix = str
	}
	sep := []byte(thousandsSep)
	n, l1, l2 := 0, len(prefix), len(sep)
	// thousands sep num
	c := (l1 - 1) / 3
	tmp := make([]byte, l2*c+l1)
	pos := len(tmp) - 1
	for i := l1 - 1; i >= 0; i, n, pos = i-1, n+1, pos-1 {
		if l2 > 0 && n > 0 && n%3 == 0 {
			for j := range sep {
				tmp[pos] = sep[l2-j-1]
				pos--
			}
		}
		tmp[pos] = prefix[i]
	}
	s := string(tmp)
	if dec > 0 {
		s += decPoint + suffix
	}
	if neg {
		s = "-" + s
	}
	return s
}

// Add Frontend Translation func
func AddTranslationV2(key string) {

	type AddTranslation struct {
		Key string `form:"key" json:"key" valid:"Required"`
	}

	processKey := strings.Split(key, ".")

	var group = processKey[0]
	// var name = processKey[1]
	name := strings.Replace(key, group+".", "", -1)
	var name2 = strings.Split(name, "_")
	var value = ""

	for _, vname := range name2 {
		res1 := strings.Contains(vname, ":")
		if !res1 {
			vname = strings.Title(vname)
		}

		if value != "" {
			value = value + " "
		}

		value = value + vname
	}

	languages, _ := models.GetLanguageList()

	for _, v := range languages {

		check, _ := models.GetTranslationByName(v.Locale, group, key)

		if check == nil {

			models.AddTranslationV2(v.Locale, group, key, value)

		}

	}
}

// MaskLeft func
func MaskLeft(s string, num int) string {
	rs := []rune(s)
	if num > 0 {
		for i := 0; i < len(rs)-num; i++ {
			rs[i] = 'x'
		}
	}

	return string(rs)
}

// GetEncryptedID func
func GetEncryptedID(code string, id int) string {
	encryptSalt := setting.Cfg.Section("custom").Key("EncryptSalt").String()
	middleIndex := len(code) / 2
	firstPart := code[0:middleIndex]
	secPart := code[middleIndex:]

	return firstPart + strconv.Itoa(id) + encryptSalt + secPart
}

func SpaceStringsBuilder(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func CompareDateTime(dateTimeStart time.Time, operand string, dateTimeEnd time.Time) bool {
	switch operand {
	case "==":
		if dateTimeStart.Format("2006-01-02 15:04:05") == dateTimeEnd.Format("2006-01-02 15:04:05") {
			return true
		}
		break
	case "<":
		if dateTimeStart.Format("2006-01-02 15:04:05") < dateTimeEnd.Format("2006-01-02 15:04:05") {
			return true
		}
		break
	case "<=":
		if dateTimeStart.Format("2006-01-02 15:04:05") <= dateTimeEnd.Format("2006-01-02 15:04:05") {
			return true
		}
		break
	case ">":
		if dateTimeStart.Format("2006-01-02 15:04:05") > dateTimeEnd.Format("2006-01-02 15:04:05") {
			return true
		}
		break
	case ">=":
		if dateTimeStart.Format("2006-01-02 15:04:05") >= dateTimeEnd.Format("2006-01-02 15:04:05") {
			return true
		}
		break
	}

	return false
}

func WeekStartDate(date time.Time) time.Time {
	// reset to the beginning hour of the days
	currentYear, currentMonth, currentDay := date.Date()
	currentLocation := date.Location()
	date = time.Date(currentYear, currentMonth, currentDay, 0, 0, 0, 0, currentLocation)

	// start logic
	offset := (int(time.Monday) - int(date.Weekday()) - 7) % 7
	result := date.Add(time.Duration(offset*24) * time.Hour)
	return result
}

func WeekEndDate(date time.Time) time.Time {
	// reset to the beginning hour of the days
	currentYear, currentMonth, currentDay := date.Date()
	currentLocation := date.Location()
	date = time.Date(currentYear, currentMonth, currentDay, 0, 0, 0, 0, currentLocation)

	// start logic
	offset := (int(time.Sunday) - int(date.Weekday()) + 7) % 7
	result := date.Add(time.Duration(offset*24) * time.Hour).Add(time.Hour*time.Duration(23) + time.Minute*time.Duration(59) + time.Second*time.Duration(59))
	return result
}

func GetWeekStartAndEndDatesWithinDateRange(dateStart, dateEnd time.Time) []map[string]time.Time {
	curWeekStart := WeekStartDate(dateStart)
	curWeekEnd := WeekEndDate(dateStart)

	// if curWeekStart is out of range, dateStart will be new curWeekStart
	if CompareDateTime(curWeekStart, "<", dateStart) {
		curWeekStart = dateStart
	}

	var arrWeeks = []map[string]time.Time{}

	// loop end if curWeekStart greater than dateEnd
	for CompareDateTime(curWeekStart, "<=", dateEnd) {
		// if curWeekEnd is out of range, dateEnd will be new curWeekEnd
		if CompareDateTime(curWeekEnd, ">", dateEnd) {
			curWeekEnd = dateEnd
		}

		arrWeeks = append(arrWeeks, map[string]time.Time{"week_start": curWeekStart, "week_end": curWeekEnd})

		curWeekStart = WeekStartDate(curWeekStart).AddDate(0, 0, +7)
		curWeekEnd = WeekEndDate(curWeekStart)
	}

	return arrWeeks
}

func GetDaysInCurrentMonth() []time.Time {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	currentDate := firstOfMonth

	var arrWeeks = []time.Time{}

	for CompareDateTime(currentDate, "<=", lastOfMonth) {
		arrWeeks = append(arrWeeks, currentDate)

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return arrWeeks
}

func GetLatestWeeks(n int) []map[string]time.Time {
	now := time.Now()

	// get current week end as dateEnd
	dateStart := WeekStartDate(now).AddDate(0, 0, -((n - 1) * 7))
	dateEnd := WeekEndDate(now)

	return GetWeekStartAndEndDatesWithinDateRange(dateStart, dateEnd)
}

func MonthStartEndDate(date time.Time) (monthStart, monthEnd time.Time) {
	// reset to the beginning hour of the days
	currentYear, currentMonth, _ := date.Date()
	currentLocation := date.Location()
	monthStart = time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	monthEnd = monthStart.AddDate(0, 1, -1)

	return monthStart, monthEnd
}

func GetMonthStartAndEndDatesWithinDateRange(dateStart, dateEnd time.Time) []map[string]time.Time {
	curMonthStart, curMonthEnd := MonthStartEndDate(dateStart)

	// if curMonthStart is out of range, dateStart will be new curMonthStart
	if CompareDateTime(curMonthStart, "<", dateStart) {
		curMonthStart = dateStart
	}

	var arrMonths = []map[string]time.Time{}

	// loop end if curMonthStart greater than dateEnd
	for CompareDateTime(curMonthStart, "<=", dateEnd) {
		// if curMonthEnd is out of range, dateEnd will be new curMonthEnd
		if CompareDateTime(curMonthEnd, ">", dateEnd) {
			curMonthEnd = dateEnd
		}

		arrMonths = append(arrMonths, map[string]time.Time{"month_start": curMonthStart, "month_end": curMonthEnd})

		curMonthStart = curMonthStart.AddDate(0, 1, 0)
		_, curMonthEnd = MonthStartEndDate(curMonthStart)
	}

	return arrMonths
}

func GetLatestMonths(n int) []map[string]time.Time {
	now := time.Now()

	// get current week end as dateEnd
	monthStart, monthEnd := MonthStartEndDate(now)
	monthStart = monthStart.AddDate(0, -(n - 1), 0)

	return GetMonthStartAndEndDatesWithinDateRange(monthStart, monthEnd)
}

func YearStartEndDate(date time.Time) (yearStart, yearEnd time.Time) {
	// reset to the beginning hour of the days
	currentYear, _, _ := date.Date()
	currentLocation := date.Location()
	yearStart = time.Date(currentYear, 1, 1, 0, 0, 0, 0, currentLocation)

	yearEnd = yearStart.AddDate(1, 0, -1)

	return yearStart, yearEnd
}

func GetYearStartAndEndDatesWithinDateRange(dateStart, dateEnd time.Time) []map[string]time.Time {
	curYearStart, curYearEnd := YearStartEndDate(dateStart)

	// if curYearStart is out of range, dateStart will be new curYearStart
	if CompareDateTime(curYearStart, "<", dateStart) {
		curYearStart = dateStart
	}

	var arrYears = []map[string]time.Time{}

	// loop end if curYearStart greater than dateEnd
	for CompareDateTime(curYearStart, "<=", dateEnd) {
		// if curYearEnd is out of range, dateEnd will be new curYearEnd
		if CompareDateTime(curYearEnd, ">", dateEnd) {
			curYearEnd = dateEnd
		}

		arrYears = append(arrYears, map[string]time.Time{"year_start": curYearStart, "year_end": curYearEnd})

		curYearStart = curYearStart.AddDate(1, 0, 0)
		_, curYearEnd = YearStartEndDate(curYearStart)
	}

	return arrYears
}

func GetLatestYears(n int) []map[string]time.Time {
	now := time.Now()

	// get current week end as dateEnd
	yearStart, yearEnd := YearStartEndDate(now)
	yearStart = yearStart.AddDate(-(n - 1), 0, 0)

	return GetYearStartAndEndDatesWithinDateRange(yearStart, yearEnd)
}

// Returns an int >= min, < max
func RandomInt(min, max int) int {
	// //fmt.Println("randomInt : ", min, max)

	// rand := rand.New(rand.NewSource(55))
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Seed(time.Now().UnixNano())

	return rand.Intn(max-min+1) + min
}

func GetStatusColorCodeByStatusCode(statusCode string) string {
	var statusColorCode = "#13B126"
	if statusCode == "V" || statusCode == "F" || statusCode == "EP" {
		statusColorCode = "#F76464"
	} else if statusCode == "P" {
		statusColorCode = "#FFA500"
	}

	return statusColorCode
}

func NumDecPlaces(v float64) int {
	s := strconv.FormatFloat(v, 'f', -1, 64)
	i := strings.IndexByte(s, '.')
	if i > -1 {
		return len(s) - i - 1
	}
	return 0
}

func FloatEquality(a, b float64) bool {
	result := big.NewFloat(a).Cmp(big.NewFloat(b))

	if result == 0 {
		return true
	}

	return false
}

func AutoTradingColorCode(code string) string {
	var colorCode string

	if code == "CFRA" {
		colorCode = "#4545EF"
	} else if code == "CIFRA" {
		colorCode = "#0797BE"
	} else if code == "SGT" {
		colorCode = "#FFB52C"
	} else if code == "MT" {
		colorCode = "#FE4931"
	} else if code == "MTD" {
		colorCode = "#a100ce"
	}

	return colorCode
}

// GetStringInBetween Returns empty string if no start string found
func GetStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	s += len(start)
	e := strings.Index(str[s:], end)
	if e == -1 {
		return
	}
	return str[s : s+e]
}

func TruncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
