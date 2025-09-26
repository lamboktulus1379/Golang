package service

import "time"

// Token captures the access credentials returned by an OAuth provider.
type Token struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expiry       time.Time
}

// UserProfile represents a subset of user information that providers typically expose.
type UserProfile struct {
	Provider string
	ID       string
	Email    string
	Name     string
	Avatar   string
}

// LoginParams captures the authorization request parameters shared by providers.
type LoginParams struct {
	State               string
	CodeChallenge       string
	CodeChallengeMethod string
}

// ExchangeParams carries inputs required when exchanging an authorization code.
type ExchangeParams struct {
	Code         string
	CodeVerifier string
}
