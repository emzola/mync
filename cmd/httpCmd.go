// Package cmd implements HTTP and GRPC sub-commands.
package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
)

type httpConfig struct {
	url  string
	verb string
}

// fetchRemoteResource returns byte data from a remote resource.
func fetchRemoteResource(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// validateConfig validates httpConfig and returns an error if verb is not GET, POST or HEAD.
func validateConfig(c httpConfig) error {
	allowedVerbs := []string{"GET", "POST", "HEAD"}
	for _, v := range allowedVerbs {
		if c.verb == v {
			return nil
		}
	}
	return ErrInvalidHTTPMethod
}

// HandleHttp handles the http command.
func HandleHttp(w io.Writer, args []string) error {
	c := httpConfig{}
	fs := flag.NewFlagSet("http", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.StringVar(&c.verb, "verb", "GET", "HTTP method")
	fs.Usage = func() {
		var usageString = `
http: A HTTP client.

http: <options> server`
		fmt.Fprint(w, usageString)
		fmt.Fprintln(w)
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Options: ")
		fs.PrintDefaults()
	}

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return ErrNoServerSpecified
	}

	// Make sure only allowed verbs are used as HTTP methods
	err = validateConfig(c)
	if err != nil {
		if errors.Is(err, ErrInvalidHTTPMethod) {
			fmt.Fprint(w, err)
		}
		return err
	}

	c.url = fs.Arg(0)

	// Fetch the remote resource
	data, err := fetchRemoteResource(c.url)
	if err != nil {
		return nil
	}

	fmt.Fprintf(w, "%s\n", data)
	return nil
}

