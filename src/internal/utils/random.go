package utils

import (
	"crypto/rand"
	"math/big"
)

const (
	// Character sets for password generation
	letterRunes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberRunes    = "0123456789"
	specialRunes   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	allRunes       = letterRunes + numberRunes + specialRunes
)

// GenerateRandomString generates a cryptographically secure random string
func GenerateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(allRunes))))
		b[i] = allRunes[num.Int64()]
	}
	return string(b)
}

// GenerateSecurePassword generates a secure password with guaranteed complexity
func GenerateSecurePassword(length int) string {
	if length < 12 {
		length = 12
	}
	
	// Ensure at least one character from each set
	password := make([]byte, length)
	
	// Add one lowercase letter
	num, _ := rand.Int(rand.Reader, big.NewInt(26))
	password[0] = letterRunes[num.Int64()]
	
	// Add one uppercase letter
	num, _ = rand.Int(rand.Reader, big.NewInt(26))
	password[1] = letterRunes[26+num.Int64()]
	
	// Add one number
	num, _ = rand.Int(rand.Reader, big.NewInt(int64(len(numberRunes))))
	password[2] = numberRunes[num.Int64()]
	
	// Add one special character
	num, _ = rand.Int(rand.Reader, big.NewInt(int64(len(specialRunes))))
	password[3] = specialRunes[num.Int64()]
	
	// Fill the rest randomly
	for i := 4; i < length; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(allRunes))))
		password[i] = allRunes[num.Int64()]
	}
	
	// Shuffle the password
	for i := length - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		password[i], password[j.Int64()] = password[j.Int64()], password[i]
	}
	
	return string(password)
}