package linear_gradient

import (
	"gonum.org/v1/gonum/mat"
)

func Perform(X mat.Dense, Y mat.Dense, Theta mat.Dense, a float64, lambda float64, num_iter int) *mat.Dense {
	r, c := X.Dims()
	m := float64(r)
	regf := regularisationFilter(c, a, lambda, m)

	var pred mat.Dense
	var diff mat.Dense
	var cost mat.Dense
	var scaled_cost mat.Dense
	X_t := X.T()

	for i := 0; i < num_iter; i++ {
		pred.Mul(&X, &Theta)
		diff.Sub(&pred, &Y)
		cost.Mul(X_t, &diff)
		scaled_cost.Scale(a/m, &cost)
		Theta.MulElem(&Theta, regf)
		Theta.Sub(&Theta, &scaled_cost)
	}

	return &Theta
}

func regularisationFilter(c int, a float64, lambda float64, m float64) *mat.Dense {
	data := make([]float64, c)
	for i := range data {
		data[i] = 1 + (a * (lambda / m))
	}
	data[0] = 1
	return mat.NewDense(c, 1, data)
}
