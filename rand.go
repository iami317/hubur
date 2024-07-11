package hubur

import (
	crand "crypto/rand"
	"github.com/thoas/go-funk"
	"io"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func RandInt(from, to int) int {
	return funk.RandomInt(from, to)
}

func RandStr(n int) string {
	return funk.RandomString(n)
}

// RandBytes generate random byte slice
func RandBytes(length int) []byte {
	if length < 1 {
		return []byte{}
	}
	b := make([]byte, length)

	if _, err := io.ReadFull(crand.Reader, b); err != nil {
		return nil
	}
	return b
}

func RandLower(n int) string {
	return funk.RandomString(n, []rune("abcdefghigklmnopqrstuvwxyz1234567890"))
}

func RandLowerLetter(n int) string {
	return funk.RandomString(n, []rune("abcdefghigklmnopqrstuvwxyz"))
}
