package domain

import "errors"

// Sentinel domain errors. Callers use errors.Is to branch on error kind.
var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrValidation   = errors.New("validation error")
	ErrInsufficient = errors.New("insufficient inventory")
	ErrCycle        = errors.New("dependency cycle detected")
)
