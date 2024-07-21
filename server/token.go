package server

import (
	"bytes"
	"math/rand/v2"
)

var (
	base58Alphabet = []rune{'1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K', 'L', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}
)

func GenServerToken() string {
	randomInt := rand.Uint64()

	var buf bytes.Buffer
	var remainder uint64
	alphabetLen := uint64(len(base58Alphabet))
	for randomInt != 0 {
		remainder = randomInt % alphabetLen
		buf.WriteRune(base58Alphabet[remainder])
		randomInt /= alphabetLen
	}

	return buf.String()
}
