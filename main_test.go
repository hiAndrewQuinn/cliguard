package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestMain(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		main()
		return
	}

	tests := []struct {
		name     string
		args     []string
		wantExit int
	}{
		{
			name:     "no args shows help",
			args:     []string{},
			wantExit: 0,
		},
		{
			name:     "help flag",
			args:     []string{"--help"},
			wantExit: 0,
		},
		{
			name:     "invalid command",
			args:     []string{"invalid"},
			wantExit: 1,
		},
		{
			name:     "validate without required flags",
			args:     []string{"validate"},
			wantExit: 1,
		},
		{
			name:     "generate without required flags",
			args:     []string{"generate"},
			wantExit: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=TestMain")
			cmd.Env = append(os.Environ(), "BE_CRASHER=1")
			cmd.Args = append(cmd.Args, tt.args...)
			err := cmd.Run()

			if e, ok := err.(*exec.ExitError); ok {
				if e.ExitCode() != tt.wantExit {
					t.Errorf("expected exit %d, got %d", tt.wantExit, e.ExitCode())
				}
			} else if tt.wantExit != 0 {
				t.Errorf("expected exit %d, got 0", tt.wantExit)
			}
		})
	}
}
