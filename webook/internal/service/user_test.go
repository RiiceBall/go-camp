package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestPasswordEncrypt(t *testing.T) {
	pwd := []byte("1234#5678abC")
	encrypted, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	assert.NoError(t, err)
	println(string(encrypted))
	err = bcrypt.CompareHashAndPassword(encrypted, pwd)
	assert.NoError(t, err)
}
