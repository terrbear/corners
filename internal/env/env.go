package env

import (
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var logLevel = log.InfoLevel
var port = 8080
var lobbyTimeout = 10

func Port() int {
	return port
}

func LobbyTimeout() int {
	return lobbyTimeout
}

func parseEnvInt(name string, defaultValue int) int {
	p, ok := os.LookupEnv("PORT")
	if ok {
		pn, err := strconv.Atoi(p)
		if err != nil {
			return pn
		} else {
			return defaultValue
		}
	}

	return defaultValue
}

func init() {
	port = parseEnvInt("PORT", 8080)
	lobbyTimeout = parseEnvInt("LOBBY_TIMEOUT", 10)

	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err == nil {
		logLevel = ll
	} else if os.Getenv("LOG_LEVEL") != "" {
		log.Infof("given invalid LOG_LEVEL=%s ; defaulting to INFO", os.Getenv("LOG_LEVEL"))
	}
	log.SetLevel(logLevel)
	log.SetReportCaller(true)
}
