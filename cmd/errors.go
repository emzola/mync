package cmd

import "errors"

var (
	ErrNoServerSpecified = errors.New("you have to specify the remote server")
	ErrInvalidHTTPMethod = errors.New("invalid HTTP method")
	ErrInvalidHTTPPostRequest = errors.New("http POST request must specify a non-empty JSON body")
	ErrInvalidHTTPPostCommand = errors.New("cannot specify both body and body-file")
	ErrInvalidHTTPCommand = errors.New("invalid HTTP command")
)