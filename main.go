package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
)

const (
	sensorPath = "/sys/bus/iio/devices/iio:device%v"

	offsetFile = "in_%v_offset"
	rawFile    = "in_%v_raw"
	scaleFile  = "in_%v_scale"

	temperature int = iota
	humidity
	accelX
	accelY
	accelZ
)

func floatFromFile(filePath string) (n float64, ok bool) {
	if data, err := ioutil.ReadFile(filePath); err != nil {
		fmt.Printf("error: %v\n", err)
		return 0, false
	} else if num, err2 := strconv.ParseFloat(strings.TrimSpace(string(data)), 64); err2 != nil {
		fmt.Printf("error: parsing %v failed with %v\n", filePath, err2)
		return 0, false
	} else {
		return num, true
	}
}

func buildFilePath(namePattern string, component int) string {
	var compS string
	var device int
	switch component {
	case temperature:
		compS = "temp"
		device = 0
	case humidity:
		compS = "humidityrelative"
		device = 0
	case accelX:
		compS = "accel_x"
		device = 1
	case accelY:
		compS = "accel_y"
		device = 1
	case accelZ:
		compS = "accel_z"
		device = 1
	}

	return path.Join(fmt.Sprintf(sensorPath, device), fmt.Sprintf(namePattern, compS))
}

func getRawOffsetScaleValue(valueKind int) (value float64, ok bool) {
	if valueKind != humidity && valueKind != temperature {
		return 0, false
	}
	offset, ok1 := floatFromFile(buildFilePath(offsetFile, valueKind))
	raw, ok2 := floatFromFile(buildFilePath(rawFile, valueKind))
	scale, ok3 := floatFromFile(buildFilePath(scaleFile, valueKind))
	return (offset + raw) * scale, ok1 && ok2 && ok3
}

func getAcceleration() (accel [3]float64, ok bool) {
	for i, valueKind := range []int{accelX, accelY, accelZ} {
		scale, ok1 := floatFromFile(buildFilePath(scaleFile, valueKind))
		raw, ok2 := floatFromFile(buildFilePath(rawFile, valueKind))
		if !ok1 || !ok2 {
			return accel, false
		}
		accel[i] = scale * raw
	}
	ok = true
	return
}

func serve(w http.ResponseWriter, r *http.Request) {
	t, _ := getRawOffsetScaleValue(temperature)
	h, _ := getRawOffsetScaleValue(humidity)
	accel, _ := getAcceleration()
	fmt.Fprintf(w, "temperature is: %v degC\n", t)
	fmt.Fprintf(w, "humidity is: %v%%\n", h)
	fmt.Fprintf(w, "acceleration is x=%v y=%v z=%v\n", accel[0], accel[1], accel[2])
}

func main() {
	http.HandleFunc("/", serve)
	if err := http.ListenAndServe(":9999", nil); err != nil {
		panic(err)
	}
}
