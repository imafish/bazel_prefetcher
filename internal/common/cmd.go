package common

import (
	"io"
	"log"
	"os"
	"os/exec"
)

func RunCmd(cmdStr string, args []string, callback func(stdout io.ReadCloser)) error {
	oldPrefix := log.Prefix()
	log.SetPrefix("common.RunCmd: ")
	defer log.SetPrefix(oldPrefix)

	cmd := exec.Command(cmdStr, args...)
	cmd.Dir = "/"

	cmd.Stderr = os.Stderr

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("failed to create stdout pipe: %v", err)
		return err
	}

	if err := cmd.Start(); err != nil {
		log.Printf("failed to start cmd `%s`, error: %v", cmdStr, err)
		return err
	}

	go callback(stdoutPipe)

	if err := cmd.Wait(); err != nil {
		log.Printf("failed to wait for cmd `%s`, error: %v", cmdStr, err)
		return err
	}

	return nil
}
