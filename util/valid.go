package util

import "regexp"

func IsValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func IsValidNRA(nra string) bool {
	re := regexp.MustCompile(`^\d{2}\.\d{2}\.\d{3}$`)
	return re.MatchString(nra)
}
