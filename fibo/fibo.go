package fibo

import (
	"fmt"
	"os"
	"time"

	"github.com/siqueiraa/util/util"
	"gopkg.in/yaml.v2"
)

var (
	actualFib   map[string]interface{}
	zone        map[string]interface{}
	currentFibo map[string]interface{}
	histFibo    []map[string]interface{}
	min_value   float64
	max_value   float64
	min_idx     int
	max_idx     int
	change      bool
	trend_fibo  string
	fiboParams  Parameters
)

type Parameters struct {
	DifPerc          float64 `yaml:"difPerc"`
	MinPercFibo      float64 `yaml:"minPercFibo"`
	MinDaysFibo      int     `yaml:"minDaysFibo"`
	CheckZoneFibo    float64 `yaml:"checkZoneFibo"`
	CheckMaxZoneFibo float64 `yaml:"checkMaxZoneFibo"`
}

type Config struct {
	Fibo Parameters `yaml:"fibo"`
}

func readConfig(file string) (Config, error) {
	var config Config

	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func checkIfvisitedFiboZone(x map[string]interface{}, index int) {
	if actualFib != nil {
		if zone["current_zone"].(float64) >= fiboParams.CheckZoneFibo && actualFib["time_min"] != x["time"] && actualFib["time_max"] != x["time"] {
			visitedValue, visitedValueExists := actualFib["visited_value"].(float64)
			if !visitedValueExists {
				if actualFib["trend_fibo"] == "Buy" {
					visitedValue = x["low"].(float64)
				} else {
					visitedValue = x["high"].(float64)
				}
				actualFib["visited_idx"] = index
				actualFib["visited_value"] = visitedValue
				actualFib["visited_time"] = x["time"]
			} else {
				if actualFib["trend_fibo"] == "Buy" {
					if x["low"].(float64) < actualFib["visited_value"].(float64) {
						actualFib["visited_idx"] = index
						actualFib["visited_value"] = x["low"]
						actualFib["visited_time"] = x["time"]
					}
				} else {
					if x["high"].(float64) > actualFib["visited_value"].(float64) {
						actualFib["visited_idx"] = index
						actualFib["visited_value"] = x["high"]
						actualFib["visited_time"] = x["time"]
					}

				}
			}

		}
	}
}

func getCurrentZone(x map[string]interface{}, index int) {
	if actualFib != nil {
		return
	}

	var dif, dif_max, dif_min, now float64

	if actualFib["trend_fibo"] == "Buy" {
		dif = actualFib["max_value"].(float64) - actualFib["min_value"].(float64)
		dif_max = actualFib["max_value"].(float64) - x["low"].(float64)
		dif_min = actualFib["max_value"].(float64) - x["high"].(float64)
		now = x["low"].(float64)
	} else {
		dif = actualFib["max_value"].(float64) - actualFib["min_value"].(float64)
		dif_max = x["high"].(float64) - actualFib["min_value"].(float64)
		dif_min = x["low"].(float64) - actualFib["min_value"].(float64)
		now = x["high"].(float64)

	}

	zone = map[string]interface{}{
		"max_value":    max_value,
		"min_value":    min_value,
		"dif_top":      dif,
		"current_dif":  dif_max,
		"now":          now,
		"time":         x["time"],
		"current_zone": (dif_max / dif) * 100,
		"max_fibo":     (dif_min / dif) * 100,
	}
}

func checkMinMax(x map[string]interface{}, index int) {
	if x["low"].(float64) < min_value {
		min_value = x["low"].(float64)
		min_idx = index
		change = true
	}

	if x["high"].(float64) > max_value {
		max_value = x["high"].(float64)
		max_idx = index
		change = true
	}

	if min_idx > 0 || max_idx > 0 {
		if min_idx < max_idx {
			trend_fibo = "Buy"
		} else {
			trend_fibo = "Sell"
		}
	}

}

func checkInsideFibo() {
	visitedValue, visitedValueExists := actualFib["visited_value"].(float64)
	lastVisitedValue, lastVisitedValueExists := actualFib["last_visited_value"].(float64)

	if visitedValueExists {

		if !lastVisitedValueExists {
			actualFib["last_visited_value"] = actualFib["visited_value"]
		} else {
			if lastVisitedValue != visitedValue {
				actualFib["last_visited_value"] = visitedValue
				return
			}
		}

		if actualFib["trend_fibo"] == "Buy" && util.CalculatePercentageDifference(visitedValue, max_value)*100 < fiboParams.MinPercFibo*1.5 {
			return
		} else if actualFib["trend_fibo"] == "Sell" && util.CalculatePercentageDifference(min_value, visitedValue)*100 < fiboParams.MinPercFibo*1.5 {
			return
		}

		histFibo = append(histFibo, actualFib)

		if actualFib["trend_fibo"] == "Buy" {
			min_value = actualFib["min_value"].(float64)
			max_value = actualFib["max_value"].(float64)
			min_idx = actualFib["visited_idx"].(int)
			max_idx = actualFib["max_idx"].(int)
		} else {
			min_value = actualFib["min_value"].(float64)
			max_value = actualFib["max_value"].(float64)
			min_idx = actualFib["min_idx"].(int)
			max_idx = actualFib["visited_idx"].(int)
		}

		actualFib = make(map[string]interface{}) // Reset the map
	}
}

func checkBigFib(x map[string]interface{}, index int) {
	if actualFib != nil {
		difPerc, _ := actualFib["dif_perc"].(float64)
		if difPerc > fiboParams.DifPerc {
			if _, ok := actualFib["min_value_big_fibo"]; !ok || ((actualFib["trend_fibo"] == "Buy" && x["high"].(float64) > actualFib["max_value_big_fibo"].(float64)) || (actualFib["trend_fibo"] == "Sell" && x["low"].(float64) < actualFib["min_value_big_fibo"].(float64))) {
				actualFib["min_value_big_fibo"] = x["low"].(float64)
				actualFib["max_value_big_fibo"] = x["high"].(float64)
				actualFib["min_idx_big_fibo"] = index
				actualFib["max_idx_big_fibo"] = index
			} else {
				if x["low"].(float64) < actualFib["min_value_big_fibo"].(float64) {
					actualFib["min_value_big_fibo"] = x["low"].(float64)
					actualFib["min_idx_big_fibo"] = index
				}
				if x["high"].(float64) > actualFib["max_value_big_fibo"].(float64) {
					actualFib["max_value_big_fibo"] = x["high"].(float64)
					actualFib["max_idx_big_fibo"] = index
				}
			}

			minValueBigFibo, _ := actualFib["min_value_big_fibo"].(float64)
			maxValueBigFibo, _ := actualFib["max_value_big_fibo"].(float64)

			if util.CalculatePercentageDifference(minValueBigFibo, maxValueBigFibo)*100 > (fiboParams.MinPercFibo*1.5) && absInt(actualFib["min_idx_big_fibo"].(int)-actualFib["max_idx_big_fibo"].(int)) >= fiboParams.MinDaysFibo {
				if actualFib["min_idx_big_fibo"].(float64) < actualFib["max_idx_big_fibo"].(float64) {
					min_idx = actualFib["min_idx_big_fibo"].(int)
					min_value = actualFib["min_value_big_fibo"].(float64)
					max_idx = index
					max_value = x["high"].(float64)
				} else {
					max_idx = actualFib["max_idx_big_fibo"].(int)
					max_value = actualFib["max_value_big_fibo"].(float64)
					min_idx = index
					min_value = x["low"].(float64)
				}

				histFibo = append(histFibo, actualFib)
				actualFib = make(map[string]interface{})

			}
		}
	}
}

func getCurrentFibo(x map[string]interface{}, df []map[string]interface{}) {
	currentFibo = map[string]interface{}{
		"now":         x["time"],
		"min_idx":     min_idx,
		"max_idx":     max_idx,
		"time_min":    df[min_idx]["time"].(time.Time),
		"min_value":   min_value,
		"time_max":    df[max_idx]["time"].(time.Time),
		"max_value":   max_value,
		"dif_perc":    util.CalculatePercentageDifference(min_value, max_value) * 100,
		"trend_fibo":  trend_fibo,
		"inside_fibo": false,
	}

}

func GetFiboHistoric(df []map[string]interface{}) []map[string]interface{} {

	configFile := "config.yaml" // Adjust this to your YAML file name
	config, err := readConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	fiboParams = config.Fibo

	for index, x := range df {
		checkBigFib(x, index)
		checkInsideFibo()
		checkMinMax(x, index)
		getCurrentZone(x, index)
		if change {
			checkIfvisitedFiboZone(x, index)
			change = false

		}
	}

}
