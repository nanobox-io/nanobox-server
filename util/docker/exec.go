package docker

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	dc "github.com/fsouza/go-dockerclient"
)

// Exec
func (d DockerUtil) ExecInContainer(container string, args ...string) ([]byte, error) {
	opts := dc.CreateExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          args,
		Container:    container,
		User:         "root",
	}
	exec, err := Client.CreateExec(opts)

	if err != nil {
		return []byte{}, err
	}
	b := &bytes.Buffer{}

	results, err := RunExec(exec, nil, b, b)

	// if 'no such file or directory' squash the error
	if strings.Contains(b.String(), "no such file or directory") {
		return b.Bytes(), nil
	}

	if err != nil {
		return b.Bytes(), err
	}

	if results.ExitCode != 0 {
		return b.Bytes(), fmt.Errorf("Bad Exit Code (%d)", results.ExitCode)
	}
	return b.Bytes(), err
}

// create a new exec object in docker
// this new exec object can then be ran.
func (d DockerUtil) CreateExec(id string, cmd []string, in, out, err bool) (*dc.Exec, error) {
	config := dc.CreateExecOptions{
		Tty:          true,
		Cmd:          cmd,
		Container:    id,
		AttachStdin:  in,
		AttachStdout: out,
		AttachStderr: err,
	}

	return Client.CreateExec(config)
}

// resize the exec.
func (d DockerUtil) ResizeExecTTY(id string, height, width int) error {
	return Client.ResizeExecTTY(id, height, width)
}

// Start the exec. This will hang until the exec exits.
func (d DockerUtil) RunExec(exec *dc.Exec, in io.Reader, out io.Writer, err io.Writer) (*dc.ExecInspect, error) {
	e := Client.StartExec(exec.ID, dc.StartExecOptions{
		Tty:          true,
		InputStream:  in,
		OutputStream: out,
		ErrorStream:  err,
		RawTerminal:  true,
	})
	if e != nil {
		return nil, e
	}
	return Client.InspectExec(exec.ID)
}
