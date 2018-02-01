package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultGetFxTbHomeDirectory(t *testing.T) {
	os.Setenv("FXTBENV_HOME", "")
	homeDir := os.ExpandEnv(`${HOME}`)
	envDir := filepath.Join(homeDir, ".fxtbenv")
	assert.Equal(t, GetFxTbHomeDirectory(), envDir)
}
