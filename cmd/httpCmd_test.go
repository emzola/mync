package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func startTestHTTPServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "this is a response")
	})	
	return httptest.NewServer(mux)
}

func TestHandleHttp(t *testing.T) {
	usageMessage := `
http: A HTTP client.

http: <options> server
	
Options:
		-output string
					File path to write the response into
		-verb string
					HTTP method (default "GET")
`

	ts := startTestHTTPServer()
	defer ts.Close()

	outputFile := filepath.Join(t.TempDir(), "file_path.out")

	testConfigs := []struct {
		args   []string
		output string
		err    error
	}{
		{
			args: []string{},
			err:  ErrNoServerSpecified,
		},
		{
			args:   []string{"-h"},
			err:    errors.New("flag: help requested"),
			output: usageMessage,
		},
		{
			args: []string{ts.URL + "/download"},
			err: nil,
			output: "this is a response\n",
		},
		{
			args:   []string{"-verb", "PUT", "http://localhost"},
			err: ErrInvalidHTTPMethod,
			output: "invalid HTTP method",
		},
		{
			args:   []string{"-verb", "GET", "-output", outputFile, ts.URL + "/download"},
			err: nil,
			output: fmt.Sprintf("Data saved to: %s\n", outputFile),
		},
	}
	byteBuf := new(bytes.Buffer)
	for _, tc := range testConfigs {
		err := HandleHttp(byteBuf, tc.args)
		if tc.err == nil && err != nil {
			t.Fatalf("Expected nil error, got %v", err)
		}

		if tc.err != nil && err.Error() != tc.err.Error() {
			t.Fatalf("Expected error %v, got %v", tc.err, err)
		}

		if len(tc.output) != 0 {
			gotOutput := byteBuf.String()
			if tc.output != gotOutput {
				t.Errorf("Expected output to be: %#v, Got: %#v", tc.output, gotOutput)
			}
		}
		byteBuf.Reset()
	}
}
