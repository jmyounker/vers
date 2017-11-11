package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/urfave/cli"
)

var version string;

func main() {
	app := cli.NewApp()
	app.Version = version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Usage: "Version file",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "init",
			Action: actionInit,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "template",
					Usage: "Choose a version file template",
				},
				cli.StringFlag{
					Name:  "rcs",
					Usage: "Override RCS specification",
				},
			},
		},
		{
			Name:   "test-config",
			Action: actionValidate,
		},
		{
			Name:   "show",
			Action: actionShow,
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "option, X",
					Usage: "Specified option",
				},
			},
		},
		{
			Name:   "data-file",
			Action: actionDataFile,
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "option, X",
					Usage: "Specified option",
				},
				cli.StringFlag{
					Name:  "data-file, o",
					Usage: "Data file",
				},
			},
		},
		{
			Name:   "bump-major",
			Action: actionBumpMajor,
		},
		{
			Name:   "bump-minor",
			Action: actionBumpMinor,
		},
		{
			Name:   "bump-release",
			Action: actionBumpRelease,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func actionInit(c *cli.Context) error {
	vf := c.GlobalString("file")
	if vf == "" {
		return errors.New("version file required")
	}
	tn := c.String("template")
	if tn == "" {
		tn = "default"
	}
	return createInitFile(vf, tn, c.String("rcs"))
}

var InitTemplates = map[string]Config{
	"default": Config{
		Data: map[string]interface{}{},
		Branches: []BranchConfig{{
			BranchPattern:   ".*",
			VersionTemplate: "{branch}.{commit-counter}",
		},
		},
		DataFileFields: []string{
			"branch",
			"commit-counter",
			"version",
		},
	},
	"semvar": Config{
		Data: map[string]interface{}{
			"major":   0,
			"minor":   0,
			"release": 1,
		},
		Branches: []BranchConfig{{
			BranchPattern:   ".*",
			VersionTemplate: "{major}.{minor}.{release}",
		},
		},
		DataFileFields: []string{
			"branch",
			"commit-counter",
			"major",
			"minor",
			"release",
			"version",
		},
	},
	"python": Config{
		Data: map[string]interface{}{
			"major":   0,
			"minor":   0,
			"release": 1,
		},
		Branches: []BranchConfig{{
			BranchPattern:   "master|trunk",
			VersionTemplate: "{major}.{minor}.{release}",
		},{
			BranchPattern:   ".*",
			VersionTemplate: "{major}.{minor}.{release}dev{commit-counter}",
		},
		},
		DataFileFields: []string{
			"branch",
			"commit-counter",
			"major",
			"minor",
			"release",
			"version",
		},
	},
}

func createInitFile(versionFile string, templateName string, rcsName string) error {
	c, ok := InitTemplates[templateName]
	if !ok {
		return fmt.Errorf("unnknown template: %s", templateName)
	}
	if rcsName == "" {
		rcs, err := GetRcs(filepath.Dir(versionFile))
		if err == nil {
			rcsName = rcs.Name()
		}
	}
	for _, df := range(RcsDataFileFields(rcsName)) {
		c.DataFileFields = append(c.DataFileFields, df)
	}
	return c.writeConfig(versionFile)
}

func RcsDataFileFields(rcsName string) []string {
	switch rcsName {
	case "git":
		return []string{
			"commit-hash",
			"commit-hash-short",
		}
	case "svn":
		return []string{
			"repo-counter",
			"repo-root",
		}
	default:
		return []string{}
	}
}

func actionShow(c *cli.Context) error {
	vf, err := GetVersionFile(c)
	if err != nil {
		return errors.New("version file required")
	}

	config, err := readConfig(vf)
	if err != nil {
		return err
	}
	if config.Branches == nil {
		return fmt.Errorf("Could not parse branches")
	}

	// Get options from command line
	opts, err := getOptions(c)
	if err != nil {
		return err
	}

	ctx := NewContext(vf, config, opts)

	// get branch from combination of supplied variables and lazy RCS
	branch, err := LookupParameter("branch", &ctx)
	if err != nil {
		return err
	}

	// locate appropriate branch config
	// if branch does not match, error
	branchConfig, err := config.getBranchConfig(branch)
	if err != nil {
		return err
	}

	format, err := ParseString(branchConfig.VersionTemplate)
	if err != nil {
		return err
	}

	// perform expansion
	version, err := format.Expand(&ctx)
	if err != nil {
		return err
	}
	fmt.Println(version)
	return nil
}

func actionDataFile(c *cli.Context) error {
	vf, err := GetVersionFile(c)
	if err != nil {
		return errors.New("version file required")
	}

	df := c.String("data-file")

	config, err := readConfig(vf)
	if err != nil {
		return err
	}
	if config.Branches == nil {
		return fmt.Errorf("Could not parse branches")
	}

	// Get options from command line
	opts, err := getOptions(c)
	if err != nil {
		return err
	}

	ctx := NewContext(vf, config, opts)

	// get branch from combination of supplied variables and lazy RCS
	branch, err := LookupParameter("branch", &ctx)
	if err != nil {
		return err
	}

	// locate appropriate branch config
	// if branch does not match, error
	branchConfig, err := config.getBranchConfig(branch)
	if err != nil {
		return err
	}

	format, err := ParseString(branchConfig.VersionTemplate)
	if err != nil {
		return err
	}

	// perform expansion
	version, err := format.Expand(&ctx)
	if err != nil {
		return err
	}
	ctx.State["version"] = version

	data := map[string]string{}
	for _, v := range ctx.Config.DataFileFields {
		value, err := LookupParameter(v, &ctx)
		if err != nil {
			return err
		}
		data[v] = value
	}

	if df == "" {
		d, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", d)
		return nil
	}
	err = writeDataFile(df, data)
	if err != nil {
		return err
	}
	return nil
}

func writeDataFile(filename string, dataFile map[string]string) error {
	data, err := json.MarshalIndent(dataFile, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0664)
}

func actionBumpMajor(c *cli.Context) error {
	vf, err := GetVersionFile(c)
	if err != nil {
		return errors.New("version file required")
	}

	config, err := readConfig(vf)
	if err != nil {
		return err
	}

	major, err := config.GetDataInt("major")
	if err != nil {
		return err
	}

	// Ensure that minor is defined
	_, err = config.GetDataInt("minor")
	if err != nil {
		return err
	}

	// Ensure that release is defined
	_, err = config.GetDataInt("release")
	if err != nil {
		return err
	}

	config.Data["major"] = major + 1
	config.Data["minor"] = 0
	config.Data["release"] = 0

	err = config.writeConfig(vf)
	if err != nil {
		return err
	}
	return nil
}

func actionBumpMinor(c *cli.Context) error {
	vf, err := GetVersionFile(c)
	if err != nil {
		return errors.New("version file required")
	}

	config, err := readConfig(vf)
	if err != nil {
		return err
	}

	minor, err := config.GetDataInt("minor")
	if err != nil {
		return err
	}

	// Ensure that release is defined in the file
	_, err = config.GetDataInt("release")
	if err != nil {
		return err
	}

	config.Data["minor"] = minor + 1
	config.Data["release"] = 0

	err = config.writeConfig(vf)
	if err != nil {
		return err
	}
	return nil
}

func actionBumpRelease(c *cli.Context) error {
	vf, err := GetVersionFile(c)
	if err != nil {
		return errors.New("version file required")
	}

	config, err := readConfig(vf)
	if err != nil {
		return err
	}

	release, err := config.GetDataInt("release")
	if err != nil {
		return err
	}

	config.Data["release"] = release + 1

	err = config.writeConfig(vf)
	if err != nil {
		return err
	}
	return nil
}

type Option struct {
	Name  string
	Value string
}

func getOptions(c *cli.Context) ([]Option, error) {
	opts := c.StringSlice("option")
	res := []Option{}
	if opts == nil {
		return res, nil
	}
	optPtrn := regexp.MustCompile("([^=]+)=(.*)")
	for _, opt := range opts {
		match := optPtrn.FindStringSubmatch(opt)
		if len(match) == 0 {
			return res, fmt.Errorf("cannot parse option '%s'", string(opt))
		}
		o := Option{
			Name:  string(match[1]),
			Value: string(match[2]),
		}
		res = append(res, o)
	}
	return res, nil
}

func GetVersionFile(c *cli.Context) (string, error) {
	rf := c.GlobalString("file")
	if rf == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("could not locate version file: %s", err.Error())
		}
		dn, err := FindInPath(ContainsVersionFile, wd)
		if err != nil {
			return "", fmt.Errorf("could not locate version file: %s", err.Error())
		}
		return filepath.Join(dn, "version.json"), nil
	}
	vf, err := filepath.Abs(filepath.Clean(rf))
	if err != nil {
		return "", err
	}
	return vf, nil
}

func actionValidate(c *cli.Context) error {
	vf := c.GlobalString("file")
	if vf == "" {
		return errors.New("version file required")
	}

	_, err := readConfig(vf)
	if err != nil {
		return err
	}
	return nil
}
