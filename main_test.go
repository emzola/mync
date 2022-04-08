package main

import (
	"bytes"
	"context"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"testing"
	"time"
)

var binaryName string

func TestMain(m *testing.M) {
	if runtime.GOOS == "windows" {
		binaryName = "mync.exe"
	} else {
		binaryName = "mync"
	}

	cmd := exec.Command("go", "build", "-o", binaryName)
	err := cmd.Run()
	if err != nil {
		os.Exit(1)
	}
	defer func() {
		err = os.Remove(binaryName)
		if err != nil {
			log.Fatalf("Error removing built binary: %v", err)
		}
	}()
	m.Run()
}

func TestSubCommandInvoke(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	binaryPath := path.Join(curDir, binaryName)
	t.Log(binaryPath)

	tests := []struct {
		args                []string
		input               string
		expectedOutputLines []string
		expectedExitCode    int
	}{
		{
			args:                []string{},
			expectedExitCode:    1,
			expectedOutputLines: []string{},
		},
		{
			args:                []string{"http"},
			expectedExitCode:    1,
			expectedOutputLines: []string{"You have to specify the remote server."},
		},
		{
			args:                []string{"http", "127.0.0.1"},
			expectedExitCode:    0,
			expectedOutputLines: []string{"Executing http command"},
		},
		{
			args:                []string{"http", "-verb", "GET", "127.0.0.1"},
			expectedExitCode:    0,
			expectedOutputLines: []string{"Executing http command"},
		},
		{
			args:                []string{"http", "-verb", "POST", "127.0.0.1"},
			expectedExitCode:    0,
			expectedOutputLines: []string{"Executing http command"},
		},
		{
			args:                []string{"http", "-method", "POST", "127.0.0.1"},
			expectedExitCode:    1,
			expectedOutputLines: []string{"flag provided but not defined: -method"},
		},
		{
			args:                []string{"grpc"},
			expectedExitCode:    1,
			expectedOutputLines: []string{"You have to specify the remote server."},
		},
		{
			args:                []string{"grpc", "127.0.0.1"},
			expectedExitCode:    0,
			expectedOutputLines: []string{"Executing grpc command"},
		},
		{
			args:                []string{"grpc", "127.0.0.1"},
			expectedExitCode:    0,
			expectedOutputLines: []string{"Executing grpc command"},
		},
	}

	byteBuf := new(bytes.Buffer)

	for _, tc := range tests {
		t.Logf("Executing: %v %v\n", binaryPath, tc.args)
		cmd := exec.CommandContext(ctx, binaryPath, tc.args...)
		cmd.Stdout = byteBuf
		if len(tc.input) != 0 {
			cmd.Stdin = strings.NewReader(tc.input)
		}
		err := cmd.Run()

		if err != nil && tc.expectedExitCode == 0 {
			t.Fatalf("Expected application to exit without an error. Got: %v", err)
		}

		if cmd.ProcessState.ExitCode() != tc.expectedExitCode {
			t.Log(byteBuf.String())
			t.Fatalf("Expected application to have exit code: %v. Got: %v", tc.expectedExitCode, cmd.ProcessState.ExitCode())
		}

		output := byteBuf.String()
		lines := strings.Split(output, "\n")
		for num := range tc.expectedOutputLines {
			if lines[num] != tc.expectedOutputLines[num] {
				t.Fatalf("Expected output line to be: %v. Got: %v", tc.expectedOutputLines, lines[num])
			}
		}
		byteBuf.Reset()
	}
}

func TestHandleCommand(t *testing.T) {
	usageMessage := `Usage: mync [http|grpc] -h

http: A HTTP client.

http: <options> server

Options: 
  -verb string
    	HTTP method (default "GET")

grpc: A gRPC client.

grpc: <options> server

Options: 
  -body string
    	Body of request
  -method string
    	Method to call
`
	testConfigs := []struct {
		args   []string
		output string
		err    error
	}{
		{
			args:   []string{},
			err:    errInvalidSubCommand,
			output: "invalid sub-command specified\n" + usageMessage,
		},
		{
			args:   []string{"-h"},
			err:    nil,
			output: usageMessage,
		},
		{
			args:   []string{"foo"},
			err:    errInvalidSubCommand,
			output: "invalid sub-command specified\n" + usageMessage,
		},
	}

	byteBuf := new(bytes.Buffer)
	for _, tc := range testConfigs {
		err := handleCommand(byteBuf, tc.args)
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
