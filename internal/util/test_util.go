/*
Copyright 2022 Brian Pursley.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"bytes"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// NewTestCommand creates a new cobra command for testing, with the specified stdin, stdout, and stderr
func NewTestCommand() (*cobra.Command, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	cmd := &cobra.Command{}
	stdin := &bytes.Buffer{}
	cmd.SetIn(stdin)
	stdout := &bytes.Buffer{}
	cmd.SetOut(stdout)
	stderr := &bytes.Buffer{}
	cmd.SetErr(stderr)
	return cmd, stdin, stdout, stderr
}

// FakeExecRunner is used to mock external program execution
type FakeExecRunner struct {
	fakeExecRuns []fakeExecRun
	RunNames     []string
	RunArgs      [][]string
}

type fakeExecRun struct {
	Stdout string
	Stderr string
	Error  error
}

// NewFakeExecRunner creates a new instance of FakeExecRunner
func NewFakeExecRunner() *FakeExecRunner {
	f := &FakeExecRunner{}
	ExecRun = f.execRun
	return f
}

// SetupRun enqueues a new mocked run of an executable, including stdout, stderr, and an error
func (f *FakeExecRunner) SetupRun(stdout, stderr string, err error) {
	f.fakeExecRuns = append(f.fakeExecRuns, fakeExecRun{
		Stdout: stdout,
		Stderr: stderr,
		Error:  err,
	})
}

// LastRunName returns the name of the last execRun
func (f *FakeExecRunner) LastRunName() string {
	return f.RunNames[len(f.RunNames)-1]
}

// LastRunArgs returns the args of the last execRun
func (f *FakeExecRunner) LastRunArgs() []string {
	return f.RunArgs[len(f.RunArgs)-1]
}

// RunCount returns the number of times execRun was called
func (f *FakeExecRunner) RunCount() int {
	return len(f.RunNames)
}

func (f *FakeExecRunner) execRun(name string, args []string, _ io.Reader, stdout, stderr io.Writer) error {
	f.RunNames = append(f.RunNames, name)
	f.RunArgs = append(f.RunArgs, args)

	if len(f.fakeExecRuns) == 0 {
		return fmt.Errorf("there are no more fake exec runs")
	}
	thisRun := f.fakeExecRuns[0]
	f.fakeExecRuns = f.fakeExecRuns[1:]

	if thisRun.Stdout != "" {
		if _, err := stdout.Write(bytes.NewBufferString(thisRun.Stdout).Bytes()); err != nil {
			return err
		}
	}
	if thisRun.Stderr != "" {
		if _, err := stderr.Write(bytes.NewBufferString(thisRun.Stderr).Bytes()); err != nil {
			return err
		}
	}
	return thisRun.Error
}
