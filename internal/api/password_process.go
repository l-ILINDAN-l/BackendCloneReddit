package api

import "golang.org/x/crypto/bcrypt"

// The password hashing function, which takes the password as an argument, returns its hashing version and error
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// The password verification function returns true if the password matches the encrypted one, otherwise false
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
