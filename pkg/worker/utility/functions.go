package utility

import "math"

func Round(num float64, precision int) float64 {
	dec := math.Pow10(precision)
	rounded := math.Round(num*dec) / dec

	return rounded
}
