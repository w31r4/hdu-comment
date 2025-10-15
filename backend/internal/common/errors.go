package common

import "errors"

var (
	// ErrEmailAlreadyUsed indicates the email has been registered.
	ErrEmailAlreadyUsed = errors.New("email already in use")
	// ErrInvalidCredentials indicates login failure.
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrReviewAlreadyProcessed indicates review status update conflict.
	ErrReviewAlreadyProcessed = errors.New("review already processed")
	// ErrInvalidRefreshToken indicates the provided refresh token is invalid or expired.
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)
