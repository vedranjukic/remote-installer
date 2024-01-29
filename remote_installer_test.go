package remote_installer

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockSession struct {
	mock.Mock
}

func (m *MockSession) Output(cmd string) ([]byte, error) {
	args := m.Called(cmd)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockSession) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockClient struct {
	mock.Mock
}

func (m *MockClient) NewSession() (SSHSession, error) {
	args := m.Called()
	return args.Get(0).(SSHSession), args.Error(1)
}

func TestDetectOs(t *testing.T) {
	expectedOutput := "Linux test 4.15.0-106-generic #107-Ubuntu SMP Thu Jun 4 11:27:52 UTC 2020 x86_64 x86_64 x86_64 GNU/Linux"

	mockSession := new(MockSession)
	mockSession.On("Output", "uname -a").Return([]byte(expectedOutput), nil)
	mockSession.On("Close").Return(nil)

	mockClient := new(MockClient)
	mockClient.On("NewSession").Return(mockSession, nil)

	installer := &RemoteInstaller{client: mockClient}
	remoteOs, err := installer.DetectOs()

	mockSession.AssertExpectations(t)
	mockClient.AssertExpectations(t)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if *remoteOs != OSLinux_64_86 {
		t.Errorf("Expected OSLinux_64_86, but got %v", remoteOs)
	}
}

func TestAgentExists(t *testing.T) {
	mockSession := new(MockSession)
	mockSession.On("Output", "test -f /usr/local/bin/daytona").Return([]byte(""), nil)
	mockSession.On("Close").Return(nil)

	mockClient := new(MockClient)
	mockClient.On("NewSession").Return(mockSession, nil)

	installer := &RemoteInstaller{client: mockClient}
	exists, err := installer.AgentExists(OSLinux_64_86)

	mockSession.AssertExpectations(t)
	mockClient.AssertExpectations(t)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !*exists {
		t.Errorf("Expected agent to exist, but it does not")
	}
}

func TestDaemonRegistered(t *testing.T) {
	mockSession := new(MockSession)
	mockSession.On("Output", "systemctl list-units --type=service | grep daytona").Return([]byte("daytona.service loaded active running"), nil)
	mockSession.On("Close").Return(nil)

	mockClient := new(MockClient)
	mockClient.On("NewSession").Return(mockSession, nil)

	installer := &RemoteInstaller{client: mockClient}
	registered, err := installer.DaemonRegistered(OSLinux_64_86)

	mockSession.AssertExpectations(t)
	mockClient.AssertExpectations(t)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !*registered {
		t.Errorf("Expected daemon to be registered, but it is not")
	}
}

func TestInstall(t *testing.T) {
	mockSession := new(MockSession)
	mockSession.On("Output", "curl -o /tmp/daytona_install.tar.gz https://example.com/linux_64_86_binary | tar -xz -C /tmp -f /tmp/daytona_install.tar.gz && mv /tmp/daytona /usr/local/bin").Return([]byte(""), nil)
	mockSession.On("Output", "chmod +x /usr/local/bin/daytona").Return([]byte(""), nil)
	mockSession.On("Close").Return(nil)

	mockClient := new(MockClient)
	mockClient.On("NewSession").Return(mockSession, nil)

	binaryUrl_linux_64_86 := "https://example.com/linux_64_86_binary"
	binaryUrl_linux_arm64 := "https://example.com/linux_arm64_binary"

	installer := &RemoteInstaller{
		client: mockClient,
		binaryUrl: map[RemoteOS]string{
			OSLinux_64_86: binaryUrl_linux_64_86,
			OSLinux_arm64: binaryUrl_linux_arm64,
		},
	}
	err := installer.Install(OSLinux_64_86)

	mockSession.AssertExpectations(t)
	mockClient.AssertExpectations(t)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
