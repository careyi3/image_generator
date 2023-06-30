package linear_regression

import (
	data_parser "github.com/careyi3/image_generator/internal/data_parser"
	file_io "github.com/careyi3/image_generator/internal/file_io"
	linear_gradient "github.com/careyi3/image_generator/internal/linear_gradient"
	"github.com/careyi3/image_generator/internal/models"
	prediction "github.com/careyi3/image_generator/internal/prediction"
)

func Run(input_file string, output_directory string, alpha float64, lambda float64, num_iters int) (*models.TrainingResults, error) {
	data, err := file_io.ReadCSVFile(input_file)

	if err != nil {
		return nil, err
	}

	ds, err := data_parser.ParseData(data)

	if err != nil {
		return nil, err
	}

	X_train := ds.TrainingSet.X
	Y_train := ds.TrainingSet.Y
	Theta := ds.Theta
	Theta = *linear_gradient.Perform(X_train, Y_train, Theta, alpha, lambda, num_iters)

	Y_pred_train := prediction.Predict(X_train, Theta)

	percentage_error_train := prediction.PercentageError(Y_train, *Y_pred_train)

	X_cv := ds.CVSet.X
	Y_cv := ds.CVSet.Y

	Y_pred_cv := prediction.Predict(X_cv, Theta)

	percentage_error_cv := prediction.PercentageError(Y_cv, *Y_pred_cv)

	X_test := ds.TestSet.X
	Y_test := ds.TestSet.Y

	Y_pred_test := prediction.Predict(X_test, Theta)

	percentage_error_test := prediction.PercentageError(Y_test, *Y_pred_test)

	err = file_io.WriteTestDataCSV(output_directory, X_test, Y_test, *Y_pred_test)
	if err != nil {
		return nil, err
	}

	results := models.TrainingResults{
		Alpha:         alpha,
		Lambda:        lambda,
		NumIterations: num_iters,
		Thetas:        Theta.RawMatrix().Data,
		TrainingError: percentage_error_train,
		CVError:       percentage_error_cv,
		TestError:     percentage_error_test,
	}

	err = file_io.WriteTrainingParamsJSON(output_directory, results)
	if err != nil {
		return nil, err
	}

	return &results, nil
}
