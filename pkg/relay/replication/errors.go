package replication

import "errors"

var IgnoreMessageError = errors.New("ignore this message and continue replication")
