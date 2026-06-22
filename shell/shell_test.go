package shell

import (
	"slices"
	"strings"
	"testing"

	"github.com/skiffos/skiff-core/config"
)

func TestBuildTargetCmdWrapsExecCommandsWithUserShell(t *testing.T) {
	s := NewShell("/home/core")
	cmd, err := s.buildTargetCmd(&config.ConfigUserShell{
		Shell: []string{"/bin/bash"},
	}, "scp -t ~/linux.iso", true)
	if err != nil {
		t.Fatal(err.Error())
	}

	expected := []string{"/bin/bash", "-c", "scp -t ~/linux.iso"}
	if !slices.Equal(cmd, expected) {
		t.Fatalf("expected %q, got %q", expected, cmd)
	}
}

func TestBuildTargetCmdRoutesSFTPSubsystemToContainerServer(t *testing.T) {
	s := NewShell("/home/core")
	cmd, err := s.buildTargetCmd(&config.ConfigUserShell{
		Shell: []string{"/bin/bash"},
	}, "/usr/libexec/sftp-server -e -l INFO", true)
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(cmd) != 7 {
		t.Fatalf("expected shim command with subsystem args, got %q", cmd)
	}
	if cmd[0] != "/bin/sh" || cmd[1] != "-c" {
		t.Fatalf("expected /bin/sh shim, got %q", cmd[:2])
	}
	if !strings.Contains(cmd[2], "/usr/lib/openssh/sftp-server") {
		t.Fatalf("expected shim to probe Debian sftp-server path, got %q", cmd[2])
	}
	expectedArgs := []string{"sftp-server", "-e", "-l", "INFO"}
	if !slices.Equal(cmd[3:], expectedArgs) {
		t.Fatalf("expected subsystem args %q, got %q", expectedArgs, cmd[3:])
	}
}

func TestBuildTargetCmdRoutesInternalSFTPToContainerServer(t *testing.T) {
	s := NewShell("/home/core")
	cmd, err := s.buildTargetCmd(&config.ConfigUserShell{}, "internal-sftp -d /home/core", true)
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(cmd) != 6 {
		t.Fatalf("expected shim command with internal-sftp args, got %q", cmd)
	}
	expectedArgs := []string{"internal-sftp", "-d", "/home/core"}
	if !slices.Equal(cmd[3:], expectedArgs) {
		t.Fatalf("expected subsystem args %q, got %q", expectedArgs, cmd[3:])
	}
}
