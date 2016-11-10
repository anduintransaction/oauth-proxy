package utils

import (
	"os"
	"os/exec"
	"runtime"

	"gottb.io/goru/errors"
)

func Fork(name string, args ...string) (*exec.Cmd, chan bool, error) {
	done := make(chan bool)
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return nil, done, errors.Wrap(err)
	}
	go func() {
		cmd.Wait()
		done <- true
	}()
	return cmd, done, nil
}

func Kill(cmd *exec.Cmd) error {
	if cmd == nil {
		return nil
	}
	if runtime.GOOS == "windows" {
		return errors.Wrap(cmd.Process.Kill())
	} else {
		return errors.Wrap(cmd.Process.Signal(os.Interrupt))
	}
}
