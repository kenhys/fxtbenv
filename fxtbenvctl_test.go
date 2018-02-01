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

func TestCustomizedGetFxTbHomeDirectory(t *testing.T) {
	envDir := "/tmp/.fxtbenv"
	os.Setenv("FXTBENV_HOME", envDir)
	assert.Equal(t, GetFxTbHomeDirectory(), envDir)
}

func TestGetFxTbProductDirectory(t *testing.T) {
	envDir := "/tmp/.fxtbenv"
	os.Setenv("FXTBENV_HOME", envDir)
	expected := "/tmp/.fxtbenv/firefox/versions/57/ja"
	assert.Equal(t, GetFxTbProductDirectory("firefox", "57", "ja"), expected)
}

func TestGetFxTbProfileDirectory(t *testing.T) {
	envDir := "/tmp/.fxtbenv"
	os.Setenv("FXTBENV_HOME", envDir)
	expected := "/tmp/.fxtbenv/firefox/profiles/57:ja@work"
	assert.Equal(t, GetFxTbProfileDirectory("firefox", "57:ja@work"), expected)
}
