package model

import "errors"

var ErrRecordNotFound = errors.New("record not found from DB")

var ErrUnsupportedSbiSchemaVersion = errors.New("unsupported SBI schema version")
