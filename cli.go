package main

import (
	"encoding/json"
	"os"

	"github.com/urfave/cli/v2"
)

var globalConfig ConfigJSON
var hasConfig bool
var saveConfig bool

func loadConfigFromFile() {

	configJson, err := os.ReadFile(DEPCHECK_DIR + "/config.json")

	if err != nil {
		return
	}

	globalConfig = ConfigJSON{}

	err = json.Unmarshal(configJson, &globalConfig)

	if err != nil {
		return
	}
	hasConfig = true
}

func createCliApp() cli.App {
	configCommands := setupConfigCLI()

	commands := []*cli.Command{
		{
			Name:  "clean",
			Usage: "Cleans all output files",
			Action: func(c *cli.Context) error {
				removeDirectory(false)
				return nil
			},
		},
		{
			Name:  "show",
			Usage: "Shows previous report",
			Action: func(c *cli.Context) error {
				openHtml()
				return nil
			},
		},
		{
			Name:    "deploy",
			Usage:   "Automatically deploy your report to netlify",
			Aliases: []string{"d"},
			Action: func(c *cli.Context) error {
				deployToNetlify()
				return nil
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "token",
					Required:    true,
					Usage:       "Netlify PAT",
					Destination: &netlifyToken,
				},
			},
		},
	}

	commands = append(commands, configCommands...)

	flags := []cli.Flag{
		&cli.BoolFlag{
			Name:        "dev",
			Aliases:     []string{"d"},
			Usage:       "Enable dev dependencies",
			Destination: &globalConfig.DevDependencies,
		},
		&cli.BoolFlag{
			Name:        "js",
			Aliases:     []string{"j"},
			Usage:       "Enable js source files",
			Destination: &globalConfig.JS,
		},
		&cli.PathFlag{
			Name:        "path",
			Aliases:     []string{"p"},
			Usage:       "Overwrite root directory",
			Destination: &globalConfig.Path,
		},
		&cli.BoolFlag{
			Name:        "log",
			Aliases:     []string{"l"},
			Usage:       "Will write logs to .depcheck.log",
			Value:       false,
			Destination: &globalConfig.Log,
		},
		&cli.StringFlag{
			Name:        "source",
			Aliases:     []string{"s"},
			Usage:       "Overwrite default sources",
			Destination: &overwriteSource,
		},
		&cli.BoolFlag{
			Name:        "report",
			Aliases:     []string{"r"},
			Usage:       "Generate report file",
			Value:       false,
			Destination: &globalConfig.Report,
		},
		&cli.BoolFlag{
			Name:        "show-versions",
			Aliases:     []string{"v"},
			Usage:       "Show conflicting versions",
			Value:       false,
			Destination: &globalConfig.Versions,
		},
		&cli.BoolFlag{
			Name:        "write-output-files",
			Aliases:     []string{"w"},
			Usage:       "This will write the esbuild output files.",
			Value:       false,
			Destination: &esbuildWrite,
		},
		&cli.StringSliceFlag{
			Name:        "externals",
			Aliases:     []string{"e"},
			Usage:       "Pass custom externals using this flag",
			Destination: &externals,
		},
		&cli.StringSliceFlag{
			Name:        "ignore-namespaces",
			Aliases:     []string{"in"},
			Usage:       "Pass namespace (@monorepo) to be ignored",
			Destination: &ignoreNameSpaces,
		},
		&cli.BoolFlag{
			Name:        "no-open",
			Aliases:     []string{"no"},
			Usage:       "Flag to prevent auto opening report in browser",
			Value:       false,
			Destination: &noOpen,
		},
		&cli.BoolFlag{
			Name:        "save-config",
			Aliases:     []string{"sc"},
			Usage:       "Flag to automatically save config from other flags",
			Value:       false,
			Destination: &saveConfig,
		},
		&cli.BoolFlag{
			Name:        "ci",
			Usage:       "Run in github actions ci mode",
			Destination: &ciMode,
		},
		&cli.StringFlag{
			Name:        "deploy",
			Usage:       "Will automatically deploy report to netlify",
			Destination: &netlifyToken,
		},
		&cli.BoolFlag{
			Name:        "browser",
			Usage:       "Will use esbuild browser platform (by default it uses node platform)",
			Value:       false,
			Destination: &globalConfig.BrowserPlatform,
		},
	}

	loadConfigFromFile()

	app := &cli.App{
		Name:                 "depp",
		EnableBashCompletion: true,
		Usage:                "Find un used packages fast",
		Flags:                flags,
		Action: func(c *cli.Context) error {

			depcheck()
			return nil
		},
		Commands: commands,
	}

	return *app
}
