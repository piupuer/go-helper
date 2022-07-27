package utils

import "golang.org/x/crypto/bcrypt"

// GenPwd The generated password is irreversible due to the use of adaptive hash algorithm
func GenPwd(str string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(str), bcrypt.DefaultCost)
	return string(hash)
}

// ComparePwd By comparing two string hashes, judge whether they are from the same plaintext
func ComparePwd(str string, pwd string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(pwd), []byte(str)); err != nil {
		return false
	}
	return true
}
