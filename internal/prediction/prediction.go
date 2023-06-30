package prediction

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

func Predict(X mat.Dense, Theta mat.Dense) *mat.Dense {
	var Y_pred mat.Dense
	Y_pred.Mul(&X, &Theta)
	return &Y_pred
}

func PercentageError(training_outputs mat.Dense, prediction mat.Dense) float64 {
	var diff mat.Dense
	diff.Sub(&training_outputs, &prediction)

	var err mat.Dense
	err.DivElem(&diff, &training_outputs)

	var square_err mat.Dense
	square_err.MulElem(&err, &err)

	var abs_err mat.Dense
	abs_err.Apply(func(i, j int, v float64) float64 { return math.Sqrt(v) }, &square_err)

	sum_square_err := mat.Sum(&abs_err)
	n := float64(prediction.RawMatrix().Rows)

	return 100 * (sum_square_err / n)
}
