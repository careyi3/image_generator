package data_parser

import (
	"encoding/json"
	"hash/fnv"
	"math/rand"
	"strconv"
	"time"

	"github.com/careyi3/image_generator/internal/models"
	"gonum.org/v1/gonum/mat"
)

func ParseData(data [][]string) (*models.ParsedInputs, error) {
	data, err := shuffle(data)
	if err != nil {
		return nil, err
	}
	rows := len(data)
	cols := len(data[0])

	Theta := createInitialTheat(cols)

	indx60 := int(float64(rows) * 0.6)
	indx80 := int(float64(rows) * 0.8)

	ts, err := createDataSet(data[0:indx60])
	if err != nil {
		return nil, err
	}

	tes, err := createDataSet(data[indx60:indx80])
	if err != nil {
		return nil, err
	}

	cvs, err := createDataSet(data[indx80:rows])
	if err != nil {
		return nil, err
	}

	pi := models.ParsedInputs{
		TrainingSet: *ts,
		CVSet:       *tes,
		TestSet:     *cvs,
		Theta:       *Theta,
	}

	return &pi, nil
}

func fetchFlatInputsAndOutputs(matrix [][]string, num_inputs int) ([]float64, []float64, error) {
	inputs := make([]float64, 0)
	outputs := make([]float64, 0)
	for _, row := range matrix {
		inputs = append(inputs, 1)
		for index, value := range row {
			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, nil, err
			}
			if index < num_inputs {
				inputs = append(inputs, num)
			} else {
				outputs = append(outputs, num)
			}
		}
	}
	return inputs, outputs, nil
}

func shuffle(matrix [][]string) ([][]string, error) {
	h := fnv.New64()
	json, err := json.MarshalIndent(matrix, "", "")
	if err != nil {
		return nil, err
	}

	h.Write(json)
	rand.Seed(int64(h.Sum64()))
	rand.Shuffle(len(matrix), func(i, j int) { matrix[i], matrix[j] = matrix[j], matrix[i] })
	return matrix, nil
}

func createDataSet(data [][]string) (*models.DataSet, error) {
	rows := len(data)
	cols := len(data[0])

	x_data, y_data, err := fetchFlatInputsAndOutputs(data, cols-1)
	if err != nil {
		return nil, err
	}

	X := mat.NewDense(rows, cols, x_data)
	Y := mat.NewDense(rows, 1, y_data)

	X, R, K := normalize(*X)

	ds := models.DataSet{
		X: *X,
		Y: *Y,
		R: *R,
		K: *K,
	}
	return &ds, nil
}

func createInitialTheat(cols int) *mat.Dense {
	theta_data := make([]float64, cols)
	for i := range theta_data {
		rand.Seed(time.Now().UnixNano())
		theta_data[i] = rand.Float64()
	}

	return mat.NewDense(cols, 1, theta_data)
}
