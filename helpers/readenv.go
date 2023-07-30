package helpers

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type ConfigService struct {
	config map[string]string
}

func LoadEnv(envFolder string) *ConfigService {

	file, err := os.Open(envFolder + "/.env")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	conf := map[string]string{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)

		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]

			if strings.HasPrefix(key, "#") {
				continue
			}

			if key == "" || value == "" {
				continue
			}

			conf[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return &ConfigService{
		config: conf,
	}
}

func (cs *ConfigService) Get(key string) (string, error) {
	res, ok := cs.config[key]
	if !ok {
		return "", fmt.Errorf("not set key %s", key)
	}

	return res, nil
}
