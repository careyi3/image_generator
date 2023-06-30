package file_io

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/careyi3/image_generator/internal/models"
	"gonum.org/v1/gonum/mat"
)

func ReadCSVFile(path string) ([][]string, error) {
	csvfile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(csvfile)

	return r.ReadAll()
}

func WriteTestDataCSV(output_directory string, inputs mat.Dense, outputs mat.Dense, pred_outputs mat.Dense) error {
	f, err := os.Create(output_directory + "/output.csv")

	if err != nil {
		return err
	}

	defer f.Close()

	r, _ := inputs.Dims()

	for i := 0; i < r; i++ {
		for _, val := range inputs.RawRowView(i) {
			_, err = f.WriteString(fmt.Sprintf("%f", val) + ",")
			if err != nil {
				return err
			}
		}
		for _, val := range outputs.RawRowView(i) {
			_, err = f.WriteString(fmt.Sprintf("%f", val) + ",")
			if err != nil {
				return err
			}
		}
		for _, val := range pred_outputs.RawRowView(i) {
			_, err = f.WriteString(fmt.Sprintf("%f", val) + "\n")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func WriteTrainingParamsJSON(output_directory string, results models.TrainingResults) error {
	file, err := json.MarshalIndent(results, "", "")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(output_directory+"/results.json", file, 0644)
	if err != nil {
		return err
	}
	return nil
}
