package service

import (
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const totpIssuer = "PicoOffice"

// GenerateTotp issues a new TOTP secret + otpauth URL
func GenerateTotp(accountName string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuer,
		AccountName: accountName,
	})
}

// VerifyTotp checks 6-digit code (±1 time step tolerance)
func VerifyTotp(secret, code string) bool {
	if secret == "" || code == "" {
		return false
	}
	return totp.Validate(code, secret)
}

// GenerateCodeForTest computes 6-digit code from secret; for smoke tests
func GenerateCodeForTest(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}
