package main

import (
	"fmt"
	"github.com/hashicorp/go-getter"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
	"strings"
)

func GetFxTbHomeDirectory() string {
	homeDir := os.ExpandEnv(`${HOME}`)
	envDir := filepath.Join(homeDir, ".fxtbenv")
	return envDir
}

func NewFxTbEnv() {
	homeDir := os.ExpandEnv(`${HOME}`)

	envDir := filepath.Join(homeDir, ".fxtbenv")
	products := []string{"firefox", "thunderbird"}
	for _, product := range products {
		entries := []string{"versions", "profiles"}
		productDir := filepath.Join(envDir, product)
		for _, entry := range entries {
			entryDir := filepath.Join(productDir, entry)
			fmt.Println("create", entryDir)
			os.MkdirAll(entryDir, 0700)
		}
	}
}

func InstallAutoconfigJsFile(installDir string) {
	jsPath := fmt.Sprintf("%s/defaults/pref/autoconfig.js", installDir)
	file, err := os.OpenFile(jsPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return
	}
	contents := []string{
		"pref(\"general.config.filename\", \"autoconfig.cfg\");",
		"pref(\"general.config.vendor\", \"autoconfig\");",
		"pref(\"general.config.obscure_value\", 0);",
	}
	file.WriteString(strings.Join(contents, "\r\n"))
}

func InstallAutoconfigCfgFile(installDir string) {
	cfgPath := fmt.Sprintf("%s/autoconfig.cfg", installDir)
	contents := []string{
		"// Disable auto update feature",
		"lockPref('app.update.auto', false);",
		"lockPref('app.update.enabled', false);",
		"lockPref('app.update.url', '');",
		"lockPref('app.update.url.override', '');",
		"lockPref('browser.search.update', false);",
	}
	file, err := os.OpenFile(cfgPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return
	}
	file.WriteString(strings.Join(contents, "\r\n"))
}

func InstallFirefox(version string) {
	base_url := "https://ftp.mozilla.org/pub/firefox/releases"
	filename := fmt.Sprintf("firefox-%s.tar.bz2", version)

	fmt.Println(filename)
	source := fmt.Sprintf("%s/%s/linux-x86_64/ja/firefox-%s.tar.bz2", base_url, version, version)
	fmt.Println(source)
	pwd, _ := os.Getwd()
	client := &getter.Client{
		Src:  source,
		Dst:  "tmp",
		Pwd:  pwd,
		Mode: getter.ClientModeDir,
	}

	if err := client.Get(); err != nil {
		fmt.Println("Error downloading: %s", err)
		os.Exit(1)
	}

	homeDir := os.ExpandEnv(`${HOME}`)
	fxDir := fmt.Sprintf("%s/.fxtbenv/firefox/versions/%s", homeDir, version)
	os.Rename("tmp/firefox", fxDir)

	InstallAutoconfigJsFile(fxDir)
	InstallAutoconfigCfgFile(fxDir)
}

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:    "install",
			Aliases: []string{"i"},
			Usage:   "Install Firefox/Thunderbird",
			Subcommands: []cli.Command{
				{
					Name:    "firefox",
					Aliases: []string{"fx"},
					Usage:   "Install Firefox",
					Action: func(c *cli.Context) error {
						NewFxTbEnv()
						InstallFirefox(c.Args().First())
						fmt.Println("install fx:", c.Args().First())
						return nil
					},
				},
				{
					Name:    "thunderbird",
					Aliases: []string{"tb"},
					Usage:   "Install Thunderbird",
					Action: func(c *cli.Context) error {
						NewFxTbEnv()
						fmt.Println("fxtb install tb:", c.Args().First())
						return nil
					},
				},
			},
		},
	}
	app.Run(os.Args)
}
