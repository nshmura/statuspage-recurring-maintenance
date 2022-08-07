package maintenance

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type StatuspageConfig struct {
	StatuspagePageId   string              `yaml:"statuspagePageId"`
	StatuspageServices []StatuspageService `yaml:"statuspageServices"`
}

type StatuspageService struct {
	Service      string   `yaml:"service"`
	ComponentIds []string `yaml:"componentIds"`
}

func (config StatuspageConfig) findComponentByServiceName(service string) *StatuspageService {

	for _, c := range config.StatuspageServices {
		if c.Service == service {
			return &c
		}
	}
	return nil
}

func loadFromFile(fileName string, data interface{}) {
	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(buf, data)
	if err != nil {
		log.Fatal(err)
	}
}
