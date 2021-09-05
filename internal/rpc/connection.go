package rpc

import (
	"net/url"

	"terrbear.io/corners/internal/env"
)

func ServerURL(playerID PlayerID) url.URL {
	return url.URL{Scheme: "ws", Host: env.GameHost(), Path: "/play/" + string(playerID)}
}
