package main

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
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
