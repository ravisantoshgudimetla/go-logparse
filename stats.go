package main

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"sort"
	"strconv"
)

func newSlice(bigSlice [][]string, title string) ([]float64, error) {
	floatValues := make([]float64, len(bigSlice)-1)
	var column int
	for i, v := range bigSlice {
		if i == 0 {
			var err error
			column, err = stringPositionInSlice(title, v)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			continue
		}
		value, _ := strconv.ParseFloat(bigSlice[i][column], 64)
		floatValues[i-1] = value
	}
	return floatValues, nil
}

// TODO: handle duplicates or none
func stringPositionInSlice(a string, list []string) (int, error) {
	for i, v := range list {
		match, _ := regexp.MatchString(a, v)
		if match {
			return i, nil
		}
	}
	return 0, fmt.Errorf("No matching headers")
}

func sum(input []float64) (sum float64) {
	for _, value := range input {
		sum += value
	}
	return
}

func mean(input []float64) (float64, error) {
	if len(input) == 0 {
		return math.NaN(), fmt.Errorf("Invalid float slice: %g", input)
	}

	return sum(input) / float64(len(input)), nil
}

func minimum(input []float64) (min float64, err error) {
	if len(input) == 0 {
		return math.NaN(), fmt.Errorf("Invalid float slice: %g", input)
	}

	min = input[0]
	for _, value := range input {
		if value < min {
			min = value
		}
	}
	return min, nil
}

func maximum(input []float64) (max float64, err error) {
	if len(input) == 0 {
		return math.NaN(), fmt.Errorf("Invalid float slice: %g", input)
	}

	max = input[0]
	for _, value := range input {
		if value > max {
			max = value
		}
	}
	return max, nil
}

func percentile(input []float64, percent float64) (percentile float64, err error) {
	if len(input) == 0 {
		return math.NaN(), fmt.Errorf("Invalid float slice: %g", input)
	}

	sort.Float64s(input)
	index := (percent / 100) * float64(len(input)-1)
	// If index happens to be a round number
	if index == float64(int64(index)) {
		i := int(index)
		return input[i], nil
	}

	// Otherwise interpolate percentile value
	k := math.Floor(index)
	f := index - k
	if int(k) >= len(input) {
		return math.NaN(), fmt.Errorf("Invalid index: %v/%v", k+1, len(input))
	}
	percentile = ((1 - f) * input[int(k)]) + (f * input[int(k)+1])
	return percentile, nil
}
