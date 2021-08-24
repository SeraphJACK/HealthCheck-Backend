package config

import (
	"flag"
	"github.com/go-yaml/yaml"
	uuid "github.com/satori/go.uuid"
	"os"
)

type Config struct {
	Token string `yaml:"Token"`
}

var Cfg = Config{Token: uuid.NewV4().String()}

func Init() error {
	path := flag.String("conf", "config.yml", "")
	flag.Parse()

	f, err := os.Open(*path)
	if err != nil {
		if os.IsNotExist(err) {
			// Generate default config
			f, err := os.Create(*path)
			if err != nil {
				return err
			}
			defer f.Close()
			return yaml.NewEncoder(f).Encode(&Cfg)
		}
		return err
	}
	defer f.Close()
	return yaml.NewDecoder(f).Decode(&Cfg)
}
