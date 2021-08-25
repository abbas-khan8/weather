package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gocarina/gocsv"
)

type Weather struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Main struct {
		Temp      float32 `json:"temp"`
		FeelsLike float32 `json:"feels_like"`
		TempMin   float32 `json:"temp_min"`
		TempMax   float32 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Wind struct {
		Speed   float32 `json:"speed"`
		Degrees int     `json:"deg"`
	} `json:"wind"`
}

type TemperatureOutput struct {
	City        string  `csv:"City"`
	Temperature float64 `csv:"Temperature"`
}

type WindOutput struct {
	City      string  `csv:"City"`
	WindSpeed float64 `csv:"Wind"`
}

func main() {
	weatherList := make([]Weather, 0)

	populateWeatherList(&weatherList)

	if len(weatherList) > 0 {
		if err := writeWeatherList(weatherList); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Process Completed")
	}
}

func populateWeatherList(weatherList *[]Weather) error {
	weatherClient := http.Client{
		Timeout: time.Second * 2,
	}

	cities := []string{
		"London", "Leeds", "Manchester", "Birmingham", "Newcastle",
		"Bristol", "Essex", "Bradford", "York", "Nottingham"}

	units := "metric"
	apiKey := "bae5f0a6b8df97353331c09833748800"

	for _, c := range cities {
		url := "https://api.openweathermap.org/data/2.5/weather"
		params := fmt.Sprintf("?q=%s&units=%s&appid=%s", c, units, apiKey)
		endpoint := url + params

		request, err := http.NewRequest(http.MethodGet, endpoint, nil)

		if err != nil {
			return fmt.Errorf("request failed! %s", err)
		}

		response, err := weatherClient.Do(request)

		if err != nil {
			return fmt.Errorf("response failed! %s", err)
		}

		if response.Body != nil {
			defer response.Body.Close()
		}

		body, err := ioutil.ReadAll(response.Body)

		if err != nil {
			return fmt.Errorf("failed to read response body! %s", err)
		}

		cityWeather := Weather{}
		jsonErr := json.Unmarshal(body, &cityWeather)

		if jsonErr != nil {
			return fmt.Errorf("failed to load JSON into Struct! %s", err)
		}

		*weatherList = append(*weatherList, cityWeather)
	}

	return nil
}

func writeWeatherList(weatherList []Weather) error {
	temperatureList, windList := extractWeatherInfo(weatherList)

	tempErr := writeTemperatures(temperatureList)

	if tempErr != nil {
		return tempErr
	}

	windError := writeWindSpeed(windList)

	if windError != nil {
		return windError
	}

	return nil
}

func writeTemperatures(temperatureList []TemperatureOutput) error {
	file, err := createCSV("highest_temperature")

	if err != nil {
		return err
	}
	defer file.Close()

	gocsv.MarshalFile(&temperatureList, file)

	return nil
}

func writeWindSpeed(windList []WindOutput) error {
	file, err := createCSV("highest_wind")

	if err != nil {
		return err
	}
	defer file.Close()

	gocsv.MarshalFile(&windList, file)

	return nil
}

func extractWeatherInfo(weatherList []Weather) ([]TemperatureOutput, []WindOutput) {
	temperatureList := make([]TemperatureOutput, len(weatherList))
	windList := make([]WindOutput, len(weatherList))

	for i, city := range weatherList {
		name := city.Name

		temperatureList[i] = TemperatureOutput{City: name, Temperature: float64(city.Main.Temp)}
		windList[i] = WindOutput{City: name, WindSpeed: float64(city.Wind.Speed)}
	}

	sort.SliceStable(temperatureList, func(i, j int) bool {
		return temperatureList[i].Temperature > temperatureList[j].Temperature
	})

	sort.SliceStable(windList, func(i, j int) bool {
		return windList[i].WindSpeed > windList[j].WindSpeed
	})

	return temperatureList[:3], windList[:3]
}

func createCSV(name string) (*os.File, error) {
	filename := fmt.Sprintf("%s.csv", name)

	file, err := os.Create(filename)

	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	return file, nil
}
