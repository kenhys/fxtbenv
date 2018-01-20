package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hashicorp/go-getter"
	version "github.com/hashicorp/go-version"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func GetFxTbHomeDirectory() string {
	envDir := os.ExpandEnv(`${FXTBENV_HOME}`)
	if envDir == "" {
		homeDir := os.ExpandEnv(`${HOME}`)
		envDir = filepath.Join(homeDir, ".fxtbenv")
	}
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
	products := []string{"firefox", "thunderbird"}
	for _, product := range products {
		targets := []string{"versions", "profiles"}
		productDir := filepath.Join(homeDir, product)
		for _, target := range targets {
			targetDir := filepath.Join(productDir, target)
			stat, err := os.Stat(targetDir)
			if err != nil {
				return false
			}
			if !stat.IsDir() {
				return false
			}
		}
	}
	return true
}

func GetSortedLabelVersions(labels []string) []string {
	versions := make([]*version.Version, len(labels))
	for i, ver := range labels {
		v, _ := version.NewVersion(ver)
		versions[i] = v
	}
	sort.Sort(version.Collection(versions))
	sorted := make([]string, len(labels))
	for i, v := range versions {
		for _, label := range labels {
			v2, _ := version.NewVersion(label)
			if v.Equal(v2) {
				sorted[i] = label
			}
		}
	}
	return sorted
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
	keys := []string{}
	for key, _ := range branches {
		if key != "devpreview" && key != "shiretoko" {
			keys = append(keys, key)
		}
	}
	keyVersions := GetSortedLabelVersions(keys)

	for _, key := range keyVersions {
		versions := GetSortedLabelVersions(branches[key])
		fmt.Print(fmt.Sprintf("%s ", strings.Split(versions[0], ".")[0]))
		fmt.Println(versions)
	}
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

func ShowInstalledProduct(products []string) {
	for _, product := range products {
		homeDir := GetFxTbHomeDirectory()
		productDir := filepath.Join(homeDir, product, "versions")
		files, err := ioutil.ReadDir(productDir)
		if err != nil {
			return
		}

		for _, file := range files {
			if file.IsDir() {
				fmt.Println(fmt.Sprintf("%s %s", product, file.Name()))
			}
		}
	}
}

func UninstallProduct(product string, version string) {
	homeDir := GetFxTbHomeDirectory()

	targetDir := filepath.Join(homeDir, product, "versions", version)
	fmt.Println(targetDir)
	if err := os.RemoveAll(targetDir); err != nil {
		fmt.Println(err)
	}
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
						if c.Bool("list") {
							GetProductVersions("thunderbird")
						}
						if c.NArg() == 0 {
							fmt.Println(fmt.Errorf("Specify Thunderbird version"))
							os.Exit(1)
						}
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
		{
			Name:    "uninstall",
			Aliases: []string{"un"},
			Usage:   "Uninstall Firefox/Thunderbird",
			Subcommands: []cli.Command{
				{
					Name:    "firefox",
					Aliases: []string{"fx"},
					Usage:   "Uninstall Firefox",
					Flags: []cli.Flag{
						cli.BoolFlag{Name: "force, f"},
					},
					Action: func(c *cli.Context) error {
						if c.Bool("force") {
						}
						if c.NArg() == 0 {
							fmt.Println(fmt.Errorf("Specify Firefox version to uninstall it."))
							os.Exit(1)
						}
						UninstallProduct("firefox", c.Args().First())
						fmt.Println("uinstall firefox:", c.Args().First())
						return nil
					},
				},
				{
					Name:    "thunderbird",
					Aliases: []string{"tb"},
					Usage:   "Install Thunderbird",
					Flags: []cli.Flag{
						cli.BoolFlag{Name: "force, f"},
					},
					Action: func(c *cli.Context) error {
						if c.Bool("force") {
						}
						if c.NArg() == 0 {
							fmt.Println(fmt.Errorf("Specify Thunderbird version to uninstall it."))
							os.Exit(1)
						}
						UninstallProduct("thunderbird", c.Args().First())
						fmt.Println("uninstall thunderbird:", c.Args().First())
						return nil
					},
				},
			},
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "List installed Firefox/Thunderbird",
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					ShowInstalledProduct([]string{"firefox", "thunderbird"})
				} else {
					ShowInstalledProduct(c.Args())
				}
				return nil
			},
		},
	}
	app.Run(os.Args)
}
