package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// config - структура для хранения конфигурации
// передаваемой в качестве аргумента коммандной строки
type config struct {
	Links        []string `json:"links"`         // массив ссылок для опроса
	SurveyPeriod int      `json:"survey_period"` // период опроса ссылок
}

// readConfig функция для чтения файла конфигурации
func readConfig(path string) (*config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var c config

	return &c, json.NewDecoder(f).Decode(&c)
}

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s <path-to-config-file>\n", os.Args[0])
		os.Exit(1)
	}

	_, err := readConfig(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}
