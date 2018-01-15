package main

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
	"github.com/hashicorp/go-getter"
)

func NewFxTbEnv() {
	homeDir := os.ExpandEnv(`${HOME}`)

	envDir := filepath.Join(homeDir, ".fxtbenv")
	products := []string{"firefox", "thunderbird"}
	for _, product := range products {
		entries := []string {"versions", "profiles"}
		productDir := filepath.Join(envDir, product)
		for _, entry := range entries {
			entryDir := filepath.Join(productDir, entry)
			fmt.Println("create", entryDir)
			os.MkdirAll(entryDir, 0700)
		}
	}
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
