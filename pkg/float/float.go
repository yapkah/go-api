package float

import (
	"math"
	"math/big"
	"strconv"
)

const prec = 200

// Add func
func Add(x, y float64) float64 {
	a := new(big.Float).SetPrec(prec).SetFloat64(x)
	b := new(big.Float).SetPrec(prec).SetFloat64(y)
	result := new(big.Float).Add(a, b)
	res, _ := strconv.ParseFloat(result.String(), 64)
	return res
}

// Sub func
func Sub(x, y float64) float64 {
	a := new(big.Float).SetPrec(prec).SetFloat64(x)
	b := new(big.Float).SetPrec(prec).SetFloat64(y)
	result := new(big.Float).Sub(a, b)
	res, _ := strconv.ParseFloat(result.String(), 64)
	return res
}

// Mul func
func Mul(x, y float64) float64 {
	a := new(big.Float).SetPrec(prec).SetFloat64(x)
	b := new(big.Float).SetPrec(prec).SetFloat64(y)
	result := new(big.Float).Mul(a, b)
	res, _ := strconv.ParseFloat(result.String(), 64)
	return res
}

// Div func
func Div(x, y float64) float64 {
	a := new(big.Float).SetPrec(prec).SetFloat64(x)
	b := new(big.Float).SetPrec(prec).SetFloat64(y)
	result := new(big.Float).Quo(a, b)
	res, _ := strconv.ParseFloat(result.String(), 64)
	return res
}

// RoundDown func
func RoundDown(value float64, d int) float64 {
	// start yee jia vers.
	// p := math.Pow(10, float64(d)) // original code
	// return math.Floor(value*p / p
	// end yee jia vers.

	// start after enahancement koo vers.
	p := 1
	for x := 0; x < d; x++ {
		if p == 0 {
			p = 1
		}
		p = p * 10
	}

	return math.Floor(value*float64(p)) / float64(p)
	// end after enahancement koo vers.
}

// RoundUp func
func RoundUp(value float64, d int) float64 {
	// p := math.Pow(10, float64(d))
	// return math.Ceil(value*p) / p
	p := 1
	for x := 0; x < d; x++ {
		if p == 0 {
			p = 1
		}
		p = p * 10
	}

	return math.Ceil(value*float64(p)) / float64(p)
}

// TrailZeroFloat func
func TrailZeroFloat(value float64, d int) (s float64) {
	var pow int = 1
	for i := 0; i < d; i++ {
		pow *= 10
	}
	value = float64(int(value*float64(pow))) / float64(pow)
	return value
}
