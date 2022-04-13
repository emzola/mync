// Package cmd implements HTTP and GRPC sub-commands.
package cmd

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

type httpConfig struct {
	url      string
	postBody string
	verb     string
}

// createRemoteResource creates data on a remote resource.
func createRemoteResource(url string, r io.Reader) ([]byte, error) {
	resp, err := http.Post(url, "application/json", r)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// fetchRemoteResource returns data from a remote resource.
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
	var validMethod bool
	allowedVerbs := []string{http.MethodGet, http.MethodPost, http.MethodHead}
	for _, v := range allowedVerbs {
		if c.verb == v {
			validMethod = true
		}
	}

	if !validMethod {
		return ErrInvalidHTTPMethod
	}

	if c.verb == http.MethodPost && len(c.postBody) == 0 {
		return ErrInvalidHTTPPostRequest
	}

	if c.verb != http.MethodPost && len(c.postBody) != 0 {
		return ErrInvalidHTTPCommand
	}

	return nil
}

// HandleHttp handles the http command.
func HandleHttp(w io.Writer, args []string) error {
	c := httpConfig{}

	var outputFile string
	var postBodyFile string
	var responseBody []byte

	fs := flag.NewFlagSet("http", flag.ContinueOnError)
	fs.SetOutput(w)
	fs.StringVar(&c.verb, "verb", "GET", "HTTP method")
	fs.StringVar(&c.postBody, "body", "", "JSON data for HTTP POST request")
	fs.StringVar(&postBodyFile, "body-file", "", "File containing JSON data for HTTP POST request")
	fs.StringVar(&outputFile, "output", "", "File path to write the response into")
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

	// Return err if -body and -body-file options are both specified
	if len(postBodyFile) != 0 && len(c.postBody) != 0 {
		return ErrInvalidHTTPPostCommand
	}

	// If there is a valid post request from -body-file, assign it to c.postbody
	if c.verb == http.MethodPost && len(postBodyFile) != 0 {
		data, err := os.ReadFile(postBodyFile)
		if err != nil {
			return nil
		}
		c.postBody = string(data)
	}

	// Validate the config to make sure only allowed verbs 
	// and appropriate sub-commands are used as HTTP methods
	err = validateConfig(c)
	if err != nil {
		if errors.Is(err, ErrInvalidHTTPMethod) || errors.Is(err, ErrInvalidHTTPPostRequest) {
			fmt.Fprintln(w, err.Error())
		}
		return err
	}

	c.url = fs.Arg(0)

	// Determine which request to make
	switch c.verb {
	case http.MethodGet:
		responseBody, err = fetchRemoteResource(c.url)
		if err != nil {
			return nil
		}
	case http.MethodPost:
		reader := bytes.NewReader([]byte(c.postBody))
		responseBody, err = createRemoteResource(c.url, reader)
		if err != nil {
			return err
		}
	}

	// if -output option is specified, create file and write data to it
	if len(outputFile) != 0 {
		f, err := os.Create(outputFile)
		if err != nil {
			return err
		}

		defer f.Close()

		_, err = f.Write(responseBody)
		if err != nil {
			return err
		}

		fmt.Fprintf(w, "Data saved to: %s\n", outputFile)
		return err
	}

	fmt.Fprintln(w, string(responseBody))
	return nil
}
