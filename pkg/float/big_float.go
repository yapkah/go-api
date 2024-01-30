package float

import (
	"math/big"
	"net/http"
	"strconv"

	"github.com/yapkah/go-api/pkg/e"
)

// BigFloat struct
type BigFloat struct {
	Number *big.Float
}

// SetString func
func SetString(number string) (*BigFloat, error) {
	bigFloatNum, ok := new(big.Float).SetPrec(prec).SetString(number)
	if !ok {
		return nil, &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.INVALID_FLOAT_STRING, Data: map[string]interface{}{"data": number}}
	}

	return &BigFloat{Number: bigFloatNum}, nil
}

// SetFloat64 func
func SetFloat64(number float64) *BigFloat {
	bigFloatNum := new(big.Float).SetPrec(prec).SetFloat64(number)
	return &BigFloat{Number: bigFloatNum}
}

// Add func
func (f *BigFloat) Add(add *BigFloat) *BigFloat {
	number := new(big.Float).Add(f.Number, add.Number)
	return &BigFloat{number}
}

// Sub func
func (f *BigFloat) Sub(sub *BigFloat) *BigFloat {
	number := new(big.Float).Sub(f.Number, sub.Number)
	return &BigFloat{number}
}

// Mul func
func (f *BigFloat) Mul(mul *BigFloat) *BigFloat {
	number := new(big.Float).Mul(f.Number, mul.Number)
	return &BigFloat{number}
}

// Div func
func (f *BigFloat) Div(div *BigFloat) *BigFloat {
	number := new(big.Float).Quo(f.Number, div.Number)
	return &BigFloat{number}
}

// Compare compare number
// -1 : f < compare
// 0 : f = compare
// 1 : f > compare
func (f *BigFloat) cmp(compare *BigFloat) int {
	return f.Number.Cmp(compare.Number)
}

// Lt lesser than
func (f *BigFloat) Lt(compare *BigFloat) bool {
	result := f.cmp(compare)
	if result == -1 {
		return true
	}
	return false
}

// Gt greater than
func (f *BigFloat) Gt(compare *BigFloat) bool {
	result := f.cmp(compare)
	if result == 1 {
		return true
	}
	return false
}

// Eq equal
func (f *BigFloat) Eq(compare *BigFloat) bool {
	result := f.cmp(compare)
	if result == 0 {
		return true
	}
	return false
}

// Lte lesser or equal
func (f *BigFloat) Lte(compare *BigFloat) bool {
	return f.Lt(compare) || f.Eq(compare)
}

// Gte greater or equal
func (f *BigFloat) Gte(compare *BigFloat) bool {
	return f.Gt(compare) || f.Eq(compare)
}

// String func
func (f *BigFloat) String(decimalSize int) string {
	return f.Number.Text('f', decimalSize)
}

// Float64 func
func (f *BigFloat) Float64() float64 {
	res, _ := strconv.ParseFloat(f.Number.String(), 64)
	return res
}

// ValueConverting func
func ValueConverting(value float64, decimal int64) *big.Int {
	valueBig := SetFloat64(value)
	decimalVal, _ := SetString(new(big.Int).Exp(big.NewInt(10), big.NewInt(decimal), nil).String())
	valueStr := valueBig.Mul(decimalVal).String(0)
	weiValue, _ := new(big.Int).SetString(valueStr, 10)
	return weiValue
}
