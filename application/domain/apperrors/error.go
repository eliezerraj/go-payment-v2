package apperrors

import (
	"errors"
)

var ErrNotFound = errors.New("item not found")
var ErrProducerKafka = errors.New("failed to connect to kafka - timeout")