package rpc

import "net/url"

const host = "corners.terrbear.io:8080"

func ServerURL(playerID PlayerID) url.URL {
	return url.URL{Scheme: "ws", Host: host, Path: "/play/" + string(playerID)}
}
