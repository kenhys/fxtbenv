package main

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func SetupTmpDir(message string) string {
	tmpDir, _ := ioutil.TempDir("", message)
	defer os.RemoveAll(tmpDir)
	return tmpDir
}

func TestDefaultGetFxTbHomeDirectory(t *testing.T) {
	os.Setenv("FXTBENV_HOME", "")
	homeDir := os.ExpandEnv(`${HOME}`)
	envDir := filepath.Join(homeDir, ".fxtbenv")
	assert.Equal(t, GetFxTbHomeDirectory(), envDir)
}

func TestCustomizedGetFxTbHomeDirectory(t *testing.T) {
	homeDir := SetupTmpDir("fxtbenv-home-directory")
	os.Setenv("FXTBENV_HOME", homeDir)
	assert.Equal(t, GetFxTbHomeDirectory(), homeDir)
}

func TestGetFxTbProductDirectory(t *testing.T) {
	homeDir := SetupTmpDir("fxtbenv-product")
	os.Setenv("FXTBENV_HOME", homeDir)
	expected := filepath.Join(homeDir, "firefox/versions/57/ja")
	assert.Equal(t, GetFxTbProductDirectory("firefox", "57", "ja"), expected)
}

func TestGetFxTbProfileDirectory(t *testing.T) {
	homeDir := SetupTmpDir("fxtbenv-profile")
	os.Setenv("FXTBENV_HOME", homeDir)
	expected := filepath.Join(homeDir, "firefox/profiles/57:ja@work")
	assert.Equal(t, GetFxTbProfileDirectory("firefox", "57:ja@work"), expected)
}

func TestIsInitializedTrue(t *testing.T) {
	homeDir := SetupTmpDir("fxtbenv-is-initialized")
	os.Setenv("FXTBENV_HOME", homeDir)
	NewFxTbEnv()
	assert.Equal(t, IsInitialized(), true)
}

func TestInstallAutoconfigJsFile(t *testing.T) {
	homeDir, _ := os.Getwd()
	os.Setenv("FXTBENV_HOME", homeDir)
	tmpDir := SetupTmpDir("fxtbenv-install-autoconfig-js")
	installDir := filepath.Join(tmpDir, "defaults/pref")
	os.MkdirAll(installDir, 0700)
	InstallAutoconfigJsFile(tmpDir)
	js := filepath.Join(installDir, "autoconfig.js")
	_, err := os.Stat(js)
	assert.Equal(t, !os.IsNotExist(err), true)
}

func TestInstallAutoconfigCfgFile(t *testing.T) {
	homeDir, _ := os.Getwd()
	os.Setenv("FXTBENV_HOME", homeDir)
	tmpDir, _ := ioutil.TempDir("", "fxtbenv-install-autoconfig-cfg")
	InstallAutoconfigCfgFile(tmpDir)
	js := filepath.Join(tmpDir, "autoconfig.cfg")
	_, err := os.Stat(js)
	assert.Equal(t, !os.IsNotExist(err), true)
}

func TestInstallDOMInspector(t *testing.T) {
	homeDir := SetupTmpDir("fxtbenv-install-dominspector")
	os.Setenv("FXTBENV_HOME", homeDir)
	version := "56"
	installDir := GetFxTbProductDirectory("firefox", version, "ja")
	InstallDOMInspector(installDir, version)
	xpi := filepath.Join(installDir, "browser/extensions/inspector@mozilla.org.xpi")
	_, err := os.Stat(xpi)
	assert.False(t, os.IsNotExist(err))
}

func TestNoInstallDOMInspector(t *testing.T) {
	homeDir := SetupTmpDir("fxtbenv-not-install-dominspector")
	os.Setenv("FXTBENV_HOME", homeDir)
	version := "57"
	installDir := GetFxTbProductDirectory("firefox", version, "ja")
	InstallDOMInspector(installDir, version)
	xpi := filepath.Join(installDir, "browser/extensions/inspector@mozilla.org.xpi")
	_, err := os.Stat(xpi)
	assert.True(t, os.IsNotExist(err))
}
