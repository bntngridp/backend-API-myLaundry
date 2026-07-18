package utils

import (
	"unicode"
)

// ValidatePasswordStrength checks if a password meets complexity requirements:
// - At least 8 characters long
// - Contains at least one uppercase letter
// - Contains at least one lowercase letter
// - Contains at least one number
// - Contains at least one special character
func ValidatePasswordStrength(password string) (bool, string) {
	var (
		hasMinLen  = len(password) >= 8
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasMinLen {
		return false, "Password must be at least 8 characters long"
	}
	if !hasUpper {
		return false, "Password must contain at least one uppercase letter (A-Z)"
	}
	if !hasLower {
		return false, "Password must contain at least one lowercase letter (a-z)"
	}
	if !hasNumber {
		return false, "Password must contain at least one number (0-9)"
	}
	if !hasSpecial {
		return false, "Password must contain at least one special character (e.g. !, @, #, $, etc.)"
	}

	return true, ""
}
