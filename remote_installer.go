package remote_installer

import (
	"fmt"
	"strings"
)

type SSHSession interface {
	Close() error
	Output(cmd string) ([]byte, error)
}

type SSHClient interface {
	NewSession() (SSHSession, error)
}

type RemoteInstaller struct {
	client    SSHClient
	binaryUrl map[RemoteOS]string
}

type RemoteOS int

const (
	OSLinux_64_86 RemoteOS = iota
	OSLinux_arm64
)

func (s *RemoteInstaller) DetectOs() (*RemoteOS, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	output, err := session.Output("uname -a")
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(string(output))
	if len(fields) < 13 {
		return nil, fmt.Errorf("unexpected output format")
	}
	cpuArch := fields[12]

	switch cpuArch {
	case "x86_64":
		arch := OSLinux_64_86
		return &arch, nil
	case "arm64":
		arch := OSLinux_arm64
		return &arch, nil
	default:
		return nil, fmt.Errorf("unexpected cpu architecture: %s", cpuArch)
	}
}

func (s *RemoteInstaller) AgentExists(os RemoteOS) (*bool, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	switch os {
	case OSLinux_64_86:
		fallthrough
	case OSLinux_arm64:
		_, err := session.Output("test -f /usr/local/bin/daytona")
		if err != nil {
			notFound := false
			return &notFound, err
		} else {
			found := true
			return &found, err
		}
	default:
		return nil, fmt.Errorf("unexpected os: %d", os)
	}
}

func (s *RemoteInstaller) DaemonRegistered(os RemoteOS) (*bool, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	switch os {
	case OSLinux_64_86:
		fallthrough
	case OSLinux_arm64:
		output, err := session.Output("systemctl list-units --type=service | grep daytona")
		if err != nil {
			return nil, err
		}

		if string(output) == "" {
			notRegistered := false
			return &notRegistered, err
		} else {
			registered := true
			return &registered, err
		}
	default:
		return nil, fmt.Errorf("unexpected os: %d", os)
	}
}

func (s *RemoteInstaller) Install(os RemoteOS) error {
	session, err := s.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	url, ok := s.binaryUrl[os]
	if !ok {
		return fmt.Errorf("url for os %d not found", os)
	}

	cmd := fmt.Sprintf("curl -o /tmp/daytona_install.tar.gz %s | tar -xz -C /tmp -f /tmp/daytona_install.tar.gz && mv /tmp/daytona /usr/local/bin", url)
	_, err = session.Output(cmd)
	if err != nil {
		return err
	}

	chmodCmd := "chmod +x /usr/local/bin/daytona"
	_, err = session.Output(chmodCmd)
	if err != nil {
		return err
	}

	return nil
}
