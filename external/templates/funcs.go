package templates

import "math/rand"

// randString is the internal function that generates a random string.
// It takes the length of the string and a string of allowed characters as parameters.
func RandString(letters string, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// randAlphaNum generates a string consisting of characters in the range 0-9, a-z, and A-Z.
func RandAlphaNum(n int) string {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return RandString(letters, n)
}

// randAlpha generates a string consisting of characters in the range a-z and A-Z.
func RandAlpha(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return RandString(letters, n)
}

// randNumeric generates a string consisting of characters in the range 0-9.
func RandNumeric(n int) string {
	const digits = "0123456789"
	return RandString(digits, n)
}

func RandInt(min, max int) int {
	return rand.Intn(max-min) + min
}
