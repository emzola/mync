package cmd

import "errors"

var ErrNoServerSpecified = errors.New("you have to specify the remote server")

var ErrInvalidHTTPMethod = errors.New("invalid HTTP method")