package mimir

import "errors"

var (
	ErrAlreadyLocked = errors.New("can't acquire lock, already locked")
)
