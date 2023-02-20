package base62

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("HI")
	encoded := Encode(467)
	isSame, _ := Decode(encoded)

	fmt.Println(strconv.FormatBool(isSame == 467))
}

const (
	ALPHABET = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	length   = uint64(len(ALPHABET))
)

func Encode(number uint64) string {

	encodedString := strings.Builder{}

	for number > 0 {
		encodedString.WriteByte(ALPHABET[number%length])
		number = number / length
	}

	return encodedString.String()
}

func Decode(encodedString string) (uint64, error) {
	number := uint64(0)

	for i, char := range encodedString {
		pos := strings.IndexRune(ALPHABET, char)

		if pos == -1 {
			return uint64(pos), errors.New("Couldn't decode " + encodedString)
		}

		number += uint64(math.Pow(float64(length), float64(i))) * uint64(pos)
	}

	return number, nil
}
