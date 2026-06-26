package domain

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrConflict      = errors.New("already exists")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrAliasTaken    = errors.New("alias already taken")
	ErrExpired       = errors.New("link expired")
	ErrInactive      = errors.New("link inactive")
)
