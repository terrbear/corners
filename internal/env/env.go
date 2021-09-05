package env

import (
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

var logLevel = log.InfoLevel
var port = 8080
var lobbyTimeout = 10
var gameHost = "corners.terrbear.io:8080"
var mapName = "og"

func Port() int {
	return port
}

func GameHost() string {
	return gameHost
}

func Map() string {
	return mapName
}

func LobbyTimeout() time.Duration {
	if lobbyTimeout <= 0 {
		return time.Millisecond
	}

	return time.Duration(lobbyTimeout) * time.Second
}

func parseEnvInt(name string, defaultValue int) int {
	val, ok := os.LookupEnv(name)
	if ok {
		log.Debug("got env var ", name, "=", val)
		num, err := strconv.Atoi(val)
		if err != nil {
			return defaultValue
		} else {
			return num
		}
	}

	return defaultValue
}

func init() {
	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err == nil {
		logLevel = ll
	} else if os.Getenv("LOG_LEVEL") != "" {
		log.Infof("given invalid LOG_LEVEL=%s ; defaulting to INFO", os.Getenv("LOG_LEVEL"))
	}
	log.SetLevel(logLevel)
	log.SetReportCaller(true)

	port = parseEnvInt("PORT", 8080)
	lobbyTimeout = parseEnvInt("LOBBY_TIMEOUT", 30)

	gh, ok := os.LookupEnv("GAME_HOST")
	if ok {
		gameHost = gh
	}

	m, ok := os.LookupEnv("MAP_NAME")
	if ok {
		mapName = m
	}
}
