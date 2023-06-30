package data_parser

import (
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
)

func normalize(input mat.Dense) (*mat.Dense, *mat.Dense, *mat.Dense) {

	r_data := make([]float64, 0)
	k_data := make([]float64, 0)

	r, c := input.Dims()
	col := make([]float64, r)
	for j := 0; j < c; j++ {
		mat.Col(col, j, &input)
		min, max := min_max(col)
		mean := stat.Mean(col, nil)

		k := -1.0
		r_inv := 0.0
		r := max - min
		if r != 0 {
			k = mean / r
			r_inv = 1 / r
		}

		for i := 0; i < c; i++ {
			if i > 0 && i == j {
				r_data = append(r_data, r_inv)
			} else {
				r_data = append(r_data, 0)
			}
		}

		k_data = append(k_data, k)

	}

	K_test := mat.NewDense(1, c, k_data)
	k_data_tmp := k_data
	for j := 0; j < r-1; j++ {
		k_data = append(k_data, k_data_tmp...)
	}

	R := mat.NewDense(c, c, r_data)
	//K := mat.NewDense(r, c, k_data)

	//norm := NormalizeBy(input, *R, *K)

	return &input, R, K_test
}

func NormalizeBy(input mat.Dense, R mat.Dense, K mat.Dense) *mat.Dense {
	var norm mat.Dense
	norm.Mul(&input, R.T())
	norm.Sub(&norm, &K)
	return &norm
}

func min_max(input []float64) (min float64, max float64) {
	min = input[0]
	max = input[0]
	for i := 0; i < len(input); i++ {
		if input[i] < min {
			min = input[i]
		}
		if input[i] > max {
			max = input[i]
		}
	}
	return min, max
}
