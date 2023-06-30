package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/careyi3/image_generator/internal/linear_regression"
	"github.com/careyi3/image_generator/internal/models"
)

func tempDir() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s/tmp", filepath.Dir(ex))
}

func wwwRoot() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s/web", filepath.Dir(ex))
}

func createDataFile(points []models.Point, degree int) (*string, error) {
	file, err := ioutil.TempFile(tempDir(), "input")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	csvw := csv.NewWriter(file)
	var data [][]string
	for _, point := range points {
		x := point.X / 100
		y := point.Y / 100

		row := make([]string, degree+1)
		for i := 0; i < degree; i++ {
			if i == 0 {
				row[i] = fmt.Sprintf("%v", x)
			} else {
				row[i] = fmt.Sprintf("%v", math.Pow(x, float64(i+1)))
			}
		}
		row[degree] = fmt.Sprintf("%v", y)
		data = append(data, row)
	}
	csvw.WriteAll(data)
	name := file.Name()
	return &name, err
}

func generateData(results models.TrainingResults, degree int) []models.Point {
	points := make([]models.Point, 100)
	x := 0.0

	for i := 0; i < 100; i++ {
		y := 0.0
		for i := 0; i < degree+1; i++ {
			if i == 0 {
				y += results.Thetas[i]
			} else if i == 1 {
				y += results.Thetas[i] * x
			} else {
				y += results.Thetas[i] * math.Pow(x, float64(i))
			}
		}
		points[i] = models.Point{X: x * 100, Y: y * 100}
		x += 0.06
	}

	return points
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Failure")
		return
	}

	var points []models.Point
	err = json.Unmarshal(body, &points)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Failure")
		return
	}

	sdegree := r.URL.Query().Get("degree")
	srate := r.URL.Query().Get("rate")
	sitrs := r.URL.Query().Get("itrs")

	degree, err := strconv.ParseInt(sdegree, 0, 32)
	rate, err := strconv.ParseFloat(srate, 32)
	itrs, err := strconv.ParseInt(sitrs, 0, 32)

	name, err := createDataFile(points, int(degree))
	defer os.Remove(*name)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Failure")
		return
	}

	results, err := linear_regression.Run(
		*name,
		tempDir(),
		rate,
		0.01,
		int(itrs),
	)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Failure")
		return
	}

	data := generateData(*results, int(degree))

	formula := "y = "
	for i := 0; i < int(degree+1); i++ {
		if i == 0 {
			formula += fmt.Sprintf("(%.2f)", results.Thetas[i])
		} else if i == 1 {
			formula += fmt.Sprintf(" + (%.2f)*x", results.Thetas[i])
		} else {
			formula += fmt.Sprintf(" + (%.2f)*x^%d", results.Thetas[i], i)
		}
	}

	response := models.ResultResponse{
		Points:        data,
		TrainingError: results.TrainingError,
		Formula:       formula,
	}
	jsonResp, err := json.Marshal(response)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Failure")
		return
	}

	fmt.Fprintf(w, string(jsonResp))
}

func main() {
	fs := http.FileServer(http.Dir(wwwRoot()))
	http.Handle("/", fs)
	http.HandleFunc("/submit", submitHandler)

	log.Println(http.ListenAndServe(":8080", nil))
}
