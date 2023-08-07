package main

import (
	"bufio"
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
	points := make([]models.Point, 20)
	x := 0.0

	for i := 0; i < 20; i++ {
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
		x += 0.3
	}

	return points
}

func generateArrowPoints(inputPoints []models.Point) []models.Point {
	points := make([]models.Point, 3*(len(inputPoints)-1))
	d := 40.0
	a := 2.5
	j := 0
	for i := 1; i < len(inputPoints); i++ {
		x2 := inputPoints[i].X
		y2 := inputPoints[i].Y
		x1 := inputPoints[i-1].X
		y1 := inputPoints[i-1].Y
		m := (y2 - y1) / (x2 - x1)
		t := math.Atan(m)
		points[j] = models.Point{X: d*math.Cos(t+a) + x2, Y: d*math.Sin(t+a) + y2}
		points[j+1] = models.Point{X: x2, Y: y2}
		points[j+2] = models.Point{X: d*math.Cos(t-a) + x2, Y: d*math.Sin(t-a) + y2}
		j += 3
	}
	return points
}

func generateGCode(points []models.Point) []string {
	commands := make([]string, 7*(len(points)/3)+14)
	commands[0] = "G90 G94"
	commands[1] = "G17"
	commands[2] = "G21"
	commands[3] = "G28 G91 Z0"
	commands[4] = "G90"
	commands[5] = "S5000 M3"

	idx := 6
	for i := 0; i < len(points); i++ {
		if i%3 == 0 {
			commands[idx] = "G0 F400"
			idx++
			commands[idx] = "Z1"
			idx++
			commands[idx] = fmt.Sprintf("X%.2f Y%.2f", points[i].X/10, points[i].Y/10)
			idx++
			commands[idx] = "Z-0.4"
			idx++
			commands[idx] = "G1 F100"
			idx++
		} else {
			commands[idx] = fmt.Sprintf("X%.2f Y%.2f", points[i].X/10, points[i].Y/10)
			idx++
		}
	}

	commands[idx] = "G0 F400"
	commands[idx+1] = "Z2"
	commands[idx+2] = "G28 G91 Z0"
	commands[idx+3] = "G90"
	commands[idx+4] = "G28 G91 X0 Y0"
	commands[idx+5] = "G90"
	commands[idx+6] = "M5"
	commands[idx+7] = "M30"

	return commands
}

func writeGCode(data []string) {
	file, err := os.OpenFile(fmt.Sprintf("%v/gcode.nc", tempDir()), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	datawriter := bufio.NewWriter(file)

	for _, line := range data {
		_, _ = datawriter.WriteString(line + "\n")
	}

	datawriter.Flush()
	file.Close()
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

	arrowData := generateArrowPoints(data)

	gcode := generateGCode(arrowData)
	writeGCode(gcode)

	response := models.ResultResponse{
		Points:        data,
		TrainingError: results.TrainingError,
		Formula:       formula,
		GCode:         gcode,
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
