package main

import (
	"fmt"
	"github.com/hashicorp/go-getter"
	"github.com/urfave/cli"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"github.com/PuerkitoBio/goquery"
)

func GetFxTbHomeDirectory() string {
	homeDir := os.ExpandEnv(`${HOME}`)
	envDir := filepath.Join(homeDir, ".fxtbenv")
	return envDir
}

func IsInitialized() bool {
	homeDir := GetFxTbHomeDirectory()
	stat, err := os.Stat(homeDir)
	if err != nil {
		return false
	}
	if !stat.IsDir() {
		return false
	}
	return true
}

func GetProductVersions(product string) []string {
	url := fmt.Sprintf("https://ftp.mozilla.org/pub/%s/releases/", product)

	doc, err := goquery.NewDocument(url)
	if err != nil {
		fmt.Print("Failed to fetch releases page")
	}
	branches := make(map[string][]string)
	doc.Find("a").Each(func(_ int, link *goquery.Selection) {
		label := strings.Replace(link.Text(), "/", "", -1)
		if !strings.ContainsAny(label, "a | b | c") && !strings.Contains(label, "..") {
			key := strings.Split(label, ".")[0]
			branches[key] = append(branches[key], label)
		}
	})
	return nil
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

func InstallProduct(product string, version string) {
	base_url := fmt.Sprintf("https://ftp.mozilla.org/pub/%s/releases", product)
	//base_url := fmt.Sprintf("http://localhost/pub/%s/releases", product)
	filename := fmt.Sprintf("%s-%s.tar.bz2", product, version)

	fmt.Println(filename)
	source := fmt.Sprintf("%s/%s/linux-x86_64/ja/%s-%s.tar.bz2", base_url, version, product, version)
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
	productDir := fmt.Sprintf("%s/.fxtbenv/%s/versions/%s", homeDir, product, version)
	os.Rename(fmt.Sprintf("tmp/%s", product), productDir)

	InstallAutoconfigJsFile(productDir)
	InstallAutoconfigCfgFile(productDir)
}

func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Usage = "Install multiple Firefox/Thunderbird and switch them."
	app.Version = "0.1.0"
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
					Flags: []cli.Flag{
						cli.BoolFlag{Name: "list, l"},
					},
					Action: func(c *cli.Context) error {
						if c.Bool("list") {
							GetProductVersions("firefox")
						}
						if c.NArg() == 0 {
							fmt.Println(fmt.Errorf("Specify Firefox version for install firefox subcommand:"))
							os.Exit(1)
						}
						if !IsInitialized() {
							NewFxTbEnv()
						}
						InstallProduct("firefox", c.Args().First())
						fmt.Println("install fx:", c.Args().First())
						return nil
					},
				},
				{
					Name:    "thunderbird",
					Aliases: []string{"tb"},
					Usage:   "Install Thunderbird",
					Flags: []cli.Flag{
						cli.BoolFlag{Name: "list, l"},
					},
					Action: func(c *cli.Context) error {
						if !IsInitialized() {
							NewFxTbEnv()
						}
						InstallProduct("thunderbird", c.Args().First())
						fmt.Println("fxtb install tb:", c.Args().First())
						return nil
					},
				},
			},
		},
	}
	app.Run(os.Args)
}
