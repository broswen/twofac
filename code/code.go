package code

import (
	"fmt"
	"math"
	"math/rand"
)

const (
	PENDING  = "PENDING"
	VERIFIED = "VERIFIED"
)

func Generate(digits int) string {
	r := rand.Intn(int(math.Pow(10, float64(digits))))
	return fmt.Sprintf("%0"+fmt.Sprintf("%d", digits)+"d", r)
}
