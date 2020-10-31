package mirror

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Command creates a exec.Cmd based on the supplied arguments.
// Stdout/Stderr are attached to those of the main process
func Command(name string, arg ...string) *exec.Cmd {
	if *verbose {
		c := append([]string{name}, arg...)
		fmt.Printf("exec: %s\n", strings.Join(c, " "))
	}

	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

// Exec runs a plain command with stdout/stderr attached to the main process
func Exec(name string, arg ...string) error {
	return Command(name, arg...).Run()
}

// ExecJson runs a command which returns JSON on it's stdout
func ExecJson(ret interface{}, name string, arg ...string) error {
	cmd := Command(name, arg...)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return err
	}

	return json.Unmarshal(out.Bytes(), ret)
}
