package util

import "math/rand"

const (
	MinListenPort = 49152
	MaxListenPort = 65535
)

func GenRandomInt(max, min int) int {
	return rand.Intn(max-min) + min
}
