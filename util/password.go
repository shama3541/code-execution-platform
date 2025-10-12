package util

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	encryptedpassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(encryptedpassword), nil
}

func VerifyPassword(hashedpassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedpassword), []byte(password))
}
