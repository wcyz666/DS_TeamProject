package utils

import(
	"math/rand"
	"time"
)

/* Usage : myrand := RandomChoice(1, 6)*/
func RandomChoice(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max - min) + min
}
