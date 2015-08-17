package radioslack

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	RedisHost         string `default:"localhost"`
	SlackClientId     string
	SlackClientSecret string
}

var config Config

func init() {
	err := envconfig.Process("rs", &config)
	if err != nil {
		log.Fatal(err.Error())
	}
}
