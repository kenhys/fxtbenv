package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/hashicorp/go-getter"
	goversion "github.com/hashicorp/go-version"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var warn = color.New(color.FgWhite, color.BgRed).SprintFunc()
var info = color.New(color.FgWhite, color.BgGreen).SprintFunc()
var debug = color.New(color.FgWhite, color.BgCyan).SprintFunc()

func GetFxTbHomeDirectory() string {
	envDir := os.ExpandEnv(`${FXTBENV_HOME}`)
	if envDir == "" {
		homeDir := os.ExpandEnv(`${HOME}`)
		envDir = filepath.Join(homeDir, ".fxtbenv")
	}
	return envDir
}

func GetFxTbProductDirectory(product string, version string, locale string) string {
	homeDir := os.ExpandEnv(`${FXTBENV_HOME}`)
	productDir := fmt.Sprintf("%s/%s/versions/%s/%s", homeDir, product, version, locale)
	return productDir
}

func GetFxTbProfileDirectory(product string, profver string) string {
	homeDir := os.ExpandEnv(`${FXTBENV_HOME}`)
	profileDir := fmt.Sprintf("%s/%s/profiles/%s", homeDir, product, profver)
	return profileDir
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
	versions := make([]*goversion.Version, len(labels))
	for i, ver := range labels {
		v, _ := goversion.NewVersion(ver)
		versions[i] = v
	}
	sort.Sort(goversion.Collection(versions))
	sorted := make([]string, len(labels))
	for i, v := range versions {
		for _, label := range labels {
			v2, _ := goversion.NewVersion(label)
			if v.Equal(v2) {
				sorted[i] = label
			}
		}
	}
	return sorted
}

func ShowProductVersions(branches map[string][]string) {
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
}

func GetProductVersions(product string) map[string][]string {
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
	return branches
}

func GetProductNightlyVersion(product string, version string) string {
	url := fmt.Sprintf("https://ftp.mozilla.org/pub/%s/nightly/latest-mozilla-central-l10n/", product)

	locale := "en-US"
	if strings.Contains(version, ":") {
		verloc := strings.Split(version, ":")
		version = verloc[0]
		locale = verloc[1]
	}

	doc, err := goquery.NewDocument(url)
	if err != nil {
		Warning("Failed to fetch nightly page")
	}
	doc.Find("a").Each(func(_ int, link *goquery.Selection) {
		filename := link.Text()
		suffix := fmt.Sprintf(".%s.linux-x86_64.tar.bz2", locale)
		if strings.HasSuffix(filename, suffix) {
			product_version := strings.Split(strings.TrimSuffix(filename, suffix), "-")
			version = product_version[1]
		}
	})
	Info("nightly", version)
	return version
}

func NewFxTbEnv() {
	envDir := os.ExpandEnv(`${FXTBENV_HOME}`)
	if envDir == "" {
		homeDir := os.ExpandEnv(`${HOME}`)
		envDir = filepath.Join(homeDir, ".fxtbenv")
	}
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
	template := fmt.Sprintf("%s/scripts/autoconfig.js", GetFxTbHomeDirectory())
	data, err := ioutil.ReadFile(template)
	if err != nil {
		return
	}
	jsPath := fmt.Sprintf("%s/defaults/pref/autoconfig.js", installDir)
	err = ioutil.WriteFile(jsPath, data, 0644)
	if err != nil {
		return
	}
}

func InstallAutoconfigCfgFile(installDir string) {
	template := fmt.Sprintf("%s/scripts/autoconfig.cfg", GetFxTbHomeDirectory())
	data, err := ioutil.ReadFile(template)
	if err != nil {
		return
	}
	cfgPath := fmt.Sprintf("%s/autoconfig.cfg", installDir)
	err = ioutil.WriteFile(cfgPath, data, 0644)
	if err != nil {
		return
	}
}

func GetReleaseProductUrl(product string, version string, useLocal bool) string {
	locale := "en-US"
	if strings.Contains(version, ":") {
		verloc := strings.SplitN(version, ":", 2)
		version = verloc[0]
		locale = verloc[1]
	}
	baseUrl := ""
	if useLocal {
		hostEnv := os.ExpandEnv(`${FXTBENV_HOST}`)
		if hostEnv != "" {
			baseUrl = fmt.Sprintf("%s/pub/%s/releases", hostEnv, product)
		} else {
			baseUrl = fmt.Sprintf("https://ftp.mozilla.org/pub/%s/releases", product)
		}
	} else {
		baseUrl = fmt.Sprintf("https://ftp.mozilla.org/pub/%s/releases", product)
	}
	filename := fmt.Sprintf("%s-%s.tar.bz2", product, version)
	url := fmt.Sprintf("%s/%s/linux-x86_64/%s/%s", baseUrl, version, locale, filename)
	return url
}

func GetNightlyProductUrl(product string, version string, useLocal bool) string {
	locale := "en-US"
	if strings.Contains(version, ":") {
		verloc := strings.SplitN(version, ":", 2)
		version = GetProductNightlyVersion(product, version)
		locale = verloc[1]
	}
	baseUrl := ""
	if useLocal {
		hostEnv := os.ExpandEnv(`${FXTBENV_HOST}`)
		if hostEnv != "" {
			baseUrl = fmt.Sprintf("%s/pub/%s/nightly/latest-mozilla-central-l10n", hostEnv, product)
		} else {
			baseUrl = fmt.Sprintf("https://ftp.mozilla.org/pub/%s/nightly/latest-mozilla-central-l10n", product)
		}
	} else {
		baseUrl = fmt.Sprintf("https://ftp.mozilla.org/pub/%s/nightly/latest-mozilla-central-l10n", product)
	}
	filename := fmt.Sprintf("%s-%s.%s.linux-x86_64.tar.bz2", product, version, locale)
	url := fmt.Sprintf("%s/%s", baseUrl, filename)
	return url
}

func GetProductSources(product string, version string) []string {
	sources := []string{}
	// version must be "nightly" or actual version string in this context
	if strings.HasPrefix(version, "nightly") {
		sources = append(sources, GetNightlyProductUrl(product, version, true))
		sources = append(sources, GetNightlyProductUrl(product, version, false))
	} else {
		sources = append(sources, GetReleaseProductUrl(product, version, true))
		sources = append(sources, GetReleaseProductUrl(product, version, false))
	}
	return sources
}

func InstallDOMInspector(productDir string, version string) {
	productVersion, _ := goversion.NewVersion(version)
	version57, _ := goversion.NewVersion("57")
	if version == "nightly" || productVersion.GreaterThan(version57) || productVersion.Equal(version57) {
		return
	}

	// Install DOM Inspector legacy Firefox (older than 57).
	// https://addons.mozilla.org/firefox/downloads/file/324966/dom_inspector-2.0.16-sm+fn+tb+fx.xpi
	// as inspector@mozilla.org.xpi
	source := "https://addons.mozilla.org/firefox/downloads/file/324966/dom_inspector-2.0.16-sm+fn+tb+fx.xpi"
	Info("Download", source)
	xpi := filepath.Join(productDir, "browser/extensions/inspector@mozilla.org.xpi")
	Info("Install to", xpi)
	DownloadFile(source, xpi)
}

func InstallProduct(product string, version string) {
	sources := GetProductSources(product, version)
	locale := "en-US"
	if strings.Contains(version, ":") {
		verloc := strings.SplitN(version, ":", 2)
		version = verloc[0]
		locale = verloc[1]
	}

	fallback := true
	for _, source := range sources {
		if !fallback {
			continue
		}
		Info("Download", source)
		pwd, _ := os.Getwd()
		client := &getter.Client{
			Src:  source,
			Dst:  "tmp",
			Pwd:  pwd,
			Mode: getter.ClientModeDir,
		}

		if err := client.Get(); err != nil {
			fmt.Println(err)
		} else {
			fallback = false
		}
	}

	productDir := GetFxTbProductDirectory(product, version, locale)
	os.MkdirAll(filepath.Dir(productDir), 0700)
	os.Rename(fmt.Sprintf("tmp/%s", product), productDir)

	// productVersion, _ := goversion.NewVersion(version)
	// version57, _ := goversion.NewVersion("57")
	// if version != "nightly" && productVersion.LessThan(version57) {
	// InstallDOMInspector(productDir, version)
	// }

	InstallAutoconfigJsFile(productDir)
	InstallAutoconfigCfgFile(productDir)
}

func ShowInstalledProduct(products []string) {
	for _, product := range products {
		homeDir := GetFxTbHomeDirectory()
		productDir := filepath.Join(homeDir, product, "versions")
		versions, err := ioutil.ReadDir(productDir)
		if err != nil {
			return
		}

		for _, version := range versions {
			if version.IsDir() {
				locales, err := ioutil.ReadDir(filepath.Join(productDir, version.Name()))
				if err != nil {
					return
				}
				for _, locale := range locales {
					if locale.IsDir() {
						fmt.Println(fmt.Sprintf("%11s %s:%s", product, version.Name(), locale.Name()))
					}
				}
			}
		}
	}
}

func UninstallProduct(product string, version string) {
	locale := "en-US"
	if strings.Contains(version, ":") {
		verloc := strings.SplitN(version, ":", 2)
		version = verloc[0]
		locale = verloc[1]
	}

	targetDir := GetFxTbProductDirectory(product, version, locale)
	fmt.Println(targetDir)
	if err := os.RemoveAll(targetDir); err != nil {
		fmt.Println(err)
	}
	homeDir := GetFxTbHomeDirectory()
	versionDir := filepath.Join(homeDir, product, "versions", version)
	if err := os.Remove(versionDir); err == nil {
		Info(fmt.Sprintf("Cleanup %s", versionDir))
	}
}

func ShowProfiles(products []string) {
	for _, product := range products {
		homeDir := GetFxTbHomeDirectory()
		productDir := filepath.Join(homeDir, product, "profiles")
		files, err := ioutil.ReadDir(productDir)
		if err != nil {
			return
		}

		for _, file := range files {
			if file.IsDir() {
				profile := ""
				if product == "firefox" {
					profile = os.ExpandEnv(`${FXTBENV_FIREFOX_PROFILE}`)
				} else if product == "thunderbird" {
					profile = os.ExpandEnv(`${FXTBENV_THUNDERBIRD_PROFILE}`)
				}
				if profile == file.Name() {
					fmt.Println(fmt.Sprintf("* %s %s", product, file.Name()))
				} else {
					fmt.Println(fmt.Sprintf("  %s %s", product, file.Name()))
				}
			}
		}
	}
}

func Warning(message string, arguments ...string) {
	fmt.Printf("%s: %s %s: ", info("fxtbenv"), warn("warning"), message)
	for _, argument := range arguments {
		fmt.Printf("%s ", argument)
	}
	fmt.Println("")
}

func Info(message string, arguments ...string) {
	fmt.Printf("%s: %s %s: ", info("fxtbenv"), info("info"), message)
	for _, argument := range arguments {
		fmt.Printf("%s ", argument)
	}
	fmt.Println("")
}

func Debug(message string, arguments ...string) {
	fmt.Printf("%s: %s %s: ", info("fxtbenv"), debug("debug"), message)
	for _, argument := range arguments {
		fmt.Printf("%s ", argument)
	}
	fmt.Println("")
}

func FxtbWErrorf(format string, value string) error {
	return fmt.Errorf("%s: %s %s: %s",
		info("fxtbenv"), warn("warning"), format, value)
}

func ParseVersionLocaleString(argument string) (string, string) {
	version := ""
	locale := "en-US"
	if strings.Contains(argument, ":") {
		r := regexp.MustCompile(`(.+):(.+)$`)
		result := r.FindAllStringSubmatch(argument, -1)
		version = result[0][1]
		locale = result[0][2]
	} else {
		version = argument
	}
	return version, locale
}

func ParseProfileString(argument string) (string, string, string, string, error) {
	message := "invalid profile argument, it must be firefox-VERSION@PROFILE or thunderbird-VERSION@PROFILE"
	if !strings.Contains(argument, "-") {
		return "", "", "", "", FxtbWErrorf(message, warn(argument))
	}
	product := ""
	profver := ""
	if strings.HasPrefix(argument, "firefox-") || strings.HasPrefix(argument, "thunderbird-") {
		arguments := strings.SplitN(argument, "-", 2)
		product = arguments[0]
		// VERSION:LOCALE@PROFILE
		profver = arguments[1]
	}
	if product != "firefox" && product != "thunderbird" {
		return product, "", "", profver, FxtbWErrorf("invalid product name", fmt.Sprintf("%s-%s", warn(product), profver))
	}
	if !strings.Contains(profver, "@") {
		return "", "", "", profver, FxtbWErrorf(message, fmt.Sprintf("%s-%s", product, warn(profver)))
	}
	version := strings.Split(profver, "@")[0]
	locale := "en-US"
	if strings.Contains(version, ":") {
		verloc := strings.Split(version, ":")
		version = verloc[0]
		locale = verloc[1]
	}
	return product, version, locale, profver, nil
}

func OpenProductDirectory(product string) {
	command := os.ExpandEnv(`${FXTBENV_FILER}`)
	if command == "" {
		command = "nautilus"
	}
	targetDir := ""
	version := ""
	locale := ""
	profile := ""
	if product == "firefox" {
		profile = os.ExpandEnv(`${FXTBENV_FIREFOX_PROFILE}`)
	} else if product == "thunderbird" {
		profile = os.ExpandEnv(`${FXTBENV_THUDERBIRD_PROFILE}`)
	} else {
	}
	r := regexp.MustCompile(`(.+):(.+)@(.+)$`)
	result := r.FindAllStringSubmatch(profile, -1)
	version = result[0][1]
	locale = result[0][2]
	targetDir = GetFxTbProductDirectory(product, version, locale)
	err := exec.Command(command, targetDir).Start()
	if err != nil {
		Warning(`Failed to launch ${command}`)
	}
}

func OpenProfileDirectory(product string) {
	command := os.ExpandEnv(`${FXTBENV_FILER}`)
	if command == "" {
		command = "nautilus"
	}
	targetDir := ""
	profile := ""
	if product == "firefox" {
		profile = os.ExpandEnv(`${FXTBENV_FIREFOX_PROFILE}`)
	} else if product == "thunderbird" {
		profile = os.ExpandEnv(`${FXTBENV_THUDERBIRD_PROFILE}`)
	} else {
	}
	targetDir = GetFxTbProfileDirectory(product, profile)
	err := exec.Command(command, targetDir).Start()
	if err != nil {
		Warning(`Failed to launch ${command}`)
	}
}

func listAction(c *cli.Context) {
	if c.NArg() == 0 {
		if c.Bool("profile") {
			ShowProfiles([]string{"firefox", "thunderbird"})
		} else {
			ShowInstalledProduct([]string{"firefox", "thunderbird"})
		}
	} else {
		if c.Bool("profile") {
			ShowProfiles(c.Args())
		} else {
			ShowInstalledProduct(c.Args())
		}
	}
}

func useAction(c *cli.Context) {
	if c.NArg() == 0 {
		ShowInstalledProduct([]string{"firefox", "thunderbird"})
	} else if c.NArg() > 1 {
		Warning("too much arguments", c.Args()...)
	} else {
		product, version, locale, profver, err := ParseProfileString(c.Args().First())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		productDir := GetFxTbProductDirectory(product, version, locale)
		stat, err := os.Stat(productDir)
		if err != nil {
			Warning(fmt.Sprintf("specified %s %s %s is not installed", product, version, locale), c.Args()...)
			os.Exit(1)
		} else {
			Info(fmt.Sprintf("%s path", product), filepath.Join(productDir, "firefox"))
		}
		profileDir := GetFxTbProfileDirectory(product, profver)
		stat, err = os.Stat(profileDir)
		if err != nil {
			if c.Bool("create") {
				Info("creating", profileDir)
				os.MkdirAll(profileDir, 0700)
			} else {
				Warning("Missing profile directory", profileDir)
				os.Exit(1)
			}
		} else {
			if !stat.IsDir() {
				Warning("invalid profile directory", c.Args()...)
				os.Exit(1)
			} else {
				Info(fmt.Sprintf("%s profile", product), profileDir)
			}
		}
	}
}

func openAction(c *cli.Context) {
	if c.NArg() == 0 {
		Warning("missing product", c.Args()...)
	} else if c.NArg() == 1 {
		if c.Bool("profile") {
			OpenProfileDirectory(c.Args().First())
		} else {
			OpenProductDirectory(c.Args().First())
		}
	} else {
		Warning("too much arguments", c.Args()...)
	}
}

func installProductAction(c *cli.Context, product string) {
	productName := strings.ToLower(product)
	if c.Bool("list") {
		ShowProductVersions(GetProductVersions(productName))
	}
	if c.NArg() == 0 {
		message := fmt.Sprintf("Specify %s version for install %s subcommand", productName, productName)
		fmt.Println(fmt.Errorf(message))
		os.Exit(1)
	}
	if !IsInitialized() {
		NewFxTbEnv()
	}
	InstallProduct(productName, c.Args().First())
	Info(fmt.Sprintf("Install %s", productName), c.Args().First())
}

func uninstallProductAction(c *cli.Context, product string) {
	if c.Bool("force") {
	}
	if c.NArg() == 0 {
		fmt.Println(fmt.Errorf(`Specify ${product} version to uninstall it.`))
		os.Exit(1)
	}
	UninstallProduct(strings.ToLower(product), c.Args().First())
	Info(`Uninstall ${strings.ToLower(product)}:`, c.Args().First())
}

func upgradeProductAction(c *cli.Context, product string) {
	if c.NArg() != 2 {
		Warning("Old version and New version must be specified")
	}
	previousVersion := c.Args().First()
	nextVersion := c.Args().Get(1)
	productName := strings.ToLower(product)
	InstallProduct(productName, nextVersion)
	UninstallProduct(productName, previousVersion)
	Info(fmt.Sprintf("Upgrade %s:%s to %s", productName, previousVersion, nextVersion))
}

func removeAction(c *cli.Context, product string) {
	profile := c.String("profile")
	if profile != "" {
		targetDir := GetFxTbProfileDirectory(strings.ToLower(product), profile)
		Info("Removing profile directory:", targetDir)
		if err := os.RemoveAll(targetDir); err != nil {
			Warning(`Failed to remove ${targetDir}`)
		}
	}
}

func resetAction(c *cli.Context, product string) {
	profileFlag := c.Bool("profile")
	if profileFlag {
		profile := ""
		if product == "firefox" {
			profile = os.ExpandEnv(`${FXTBENV_FIREFOX_PROFILE}`)
			Info("Reset Firefox profile", profile)
		} else if product == "thunderbird" {
			profile = os.ExpandEnv(`${FXTBENV_THUNDERBIRD_PROFILE}`)
			Info("Reset Thunderbird profile", profile)
		} else {
		}
		targetDir := GetFxTbProfileDirectory(product, profile)
		Info("Removing profile directory", targetDir)
		if err := os.RemoveAll(targetDir); err != nil {
			Warning(`Failed to remove ${targetDir}`)
		}
		Info("Creating profile directory", targetDir)
		os.MkdirAll(targetDir, 0700)
	} else {
		// Reset installed product
	}
}

func DownloadFile(source string, destination string) {
	Debug("source", source)
	Debug("destination", destination)
	pwd, _ := os.Getwd()
	url := fmt.Sprintf("%s?archive=false", source)
	Info("Download", url)
	client := &getter.Client{
		Src:  url,
		Dst:  destination,
		Pwd:  pwd,
		Mode: getter.ClientModeFile,
	}
	if err := client.Get(); err != nil {
		Warning("Failed to download", destination)
		fmt.Println(err)
	}
}

func mirrorAction(c *cli.Context, product string) {
	rootDir := ""
	root := os.ExpandEnv(`${FXTBENV_MIRROR_ROOT}`)
	if root != "" {
		rootDir = root
	}
	root = c.String("root")
	if root != "" {
		rootDir = root
	}
	os.MkdirAll(rootDir, 0700)
	// Fetch version list
	Debug("Product", product)
	branches := GetProductVersions(strings.ToLower(product))

	for _, verloc := range c.Args() {
		version, locale := ParseVersionLocaleString(verloc)
		versions := []string{}
		for key, _ := range branches {
			Debug("Mirror key", key)
			if key == version {
				Debug("Match key", key)
				versions = branches[key]
			}
		}
		if len(versions) == 0 {
			Debug("Mirror oneshot", version)
			versions = append(versions, version)
		}
		for _, version := range versions {
			path := fmt.Sprintf("pub/%s/releases/%s/linux-x86_64/%s", strings.ToLower(product), version, locale)
			filename := fmt.Sprintf("%s-%s.tar.bz2", strings.ToLower(product), version)
			source := fmt.Sprintf("https://ftp.mozilla.org/%s/%s", path, filename)
			targetDir := filepath.Join(rootDir, path)
			os.MkdirAll(targetDir, 0700)
			destination := filepath.Join(targetDir, filename)
			DownloadFile(source, destination)
		}
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "fxtbenv"
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
						installProductAction(c, "Firefox")
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
						installProductAction(c, "Thunderbird")
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
						uninstallProductAction(c, "Firefox")
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
						uninstallProductAction(c, "Thunderbird")
						return nil
					},
				},
			},
		},
		{
			Name:    "upgrade",
			Aliases: []string{"up"},
			Usage:   "Upgrade Firefox/Thunderbird",
			Subcommands: []cli.Command{
				{
					Name:    "firefox",
					Aliases: []string{"fx"},
					Usage:   "Upgrade Firefox",
					Action: func(c *cli.Context) error {
						upgradeProductAction(c, "Firefox")
						return nil
					},
				},
				{
					Name:    "thunderbird",
					Aliases: []string{"tb"},
					Usage:   "Upgrade Thunderbird",
					Action: func(c *cli.Context) error {
						upgradeProductAction(c, "Thunderbird")
						return nil
					},
				},
			},
		},
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "List installed Firefox/Thunderbird",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "profile, p"},
			},
			Action: func(c *cli.Context) error {
				listAction(c)
				return nil
			},
		},
		{
			Name:    "use",
			Aliases: []string{"u"},
			Usage:   "Switch to specific profile",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "create, c"},
			},
			Action: func(c *cli.Context) error {
				useAction(c)
				return nil
			},
		},
		{
			Name:    "open",
			Aliases: []string{"o"},
			Usage:   "Open specific directory",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "profile, p"},
			},
			Action: func(c *cli.Context) error {
				openAction(c)
				return nil
			},
		},
		{
			Name:    "remove",
			Aliases: []string{"r"},
			Usage:   "Remove specific product",
			Subcommands: []cli.Command{
				{
					Name:    "firefox",
					Aliases: []string{"fx"},
					Usage:   "Remove Firefox",
					Flags: []cli.Flag{
						cli.BoolFlag{Name: "force, f"},
						cli.StringFlag{Name: "profile, p"},
					},
					Action: func(c *cli.Context) error {
						removeAction(c, "Firefox")
						return nil
					},
				},
				{
					Name:    "thunderbird",
					Aliases: []string{"tb"},
					Usage:   "Remove Thunderbird",
					Flags: []cli.Flag{
						cli.BoolFlag{Name: "force, f"},
						cli.StringFlag{Name: "profile, p"},
					},
					Action: func(c *cli.Context) error {
						removeAction(c, "Thunderbird")
						return nil
					},
				},
			},
		},
		{
			Name:    "mirror",
			Aliases: []string{"m"},
			Usage:   "Mirror specific product",
			Subcommands: []cli.Command{
				{
					Name:    "firefox",
					Aliases: []string{"fx"},
					Usage:   "Mirror Firefox",
					Flags: []cli.Flag{
						cli.BoolFlag{Name: "force, f"},
						cli.StringFlag{Name: "root, r"},
					},
					Action: func(c *cli.Context) error {
						mirrorAction(c, "Firefox")
						return nil
					},
				},
				{
					Name:    "thunderbird",
					Aliases: []string{"tb"},
					Usage:   "Mirror Thunderbird",
					Flags: []cli.Flag{
						cli.BoolFlag{Name: "force, f"},
						cli.StringFlag{Name: "root, r"},
					},
					Action: func(c *cli.Context) error {
						mirrorAction(c, "Thunderbird")
						return nil
					},
				},
			},
		},
		{
			Name:    "reset",
			Aliases: []string{"m"},
			Usage:   "Reset specific product",
			Subcommands: []cli.Command{
				{
					Name:    "firefox",
					Aliases: []string{"fx"},
					Usage:   "Mirror Firefox",
					Flags: []cli.Flag{
						cli.BoolFlag{Name: "profile, p"},
					},
					Action: func(c *cli.Context) error {
						resetAction(c, "firefox")
						return nil
					},
				},
				{
					Name:    "thunderbird",
					Aliases: []string{"tb"},
					Usage:   "Mirror Thunderbird",
					Flags: []cli.Flag{
						cli.BoolFlag{Name: "profile, p"},
					},
					Action: func(c *cli.Context) error {
						resetAction(c, "thunderbird")
						return nil
					},
				},
			},
		},
	}
	app.Run(os.Args)
}
