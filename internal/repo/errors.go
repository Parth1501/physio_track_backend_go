package repo

import "errors"

var ErrNotFound = errors.New("not found")
var ErrForbidden = errors.New("forbidden")
var ErrExceedsPending = errors.New("amount exceeds pending total")
