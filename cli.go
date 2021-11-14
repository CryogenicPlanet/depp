package main

import (
	"github.com/urfave/cli/v2"
)

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
			Destination: &globalConfig.Report,
		},
		&cli.BoolFlag{
			Name:        "show-versions",
			Aliases:     []string{"v"},
			Usage:       "Show conflicting versions",
			Destination: &globalConfig.Versions,
		},
		&cli.BoolFlag{
			Name:        "write-output-files",
			Aliases:     []string{"w"},
			Usage:       "This will write the esbuild output files.",
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
			Destination: &noOpen,
		},
		&cli.BoolFlag{
			Name:        "save-config",
			Aliases:     []string{"sc"},
			Usage:       "Flag to automatically save config from other flags",
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
			Destination: &globalConfig.BrowserPlatform,
		},
		&cli.StringSliceFlag{
			Name:        "ignore-path",
			Aliases:     []string{"ip"},
			Usage:       "A glob pattern of files to be ignored",
			Destination: &ignorePaths,
		},
	}

	app := &cli.App{
		Name:                 "depp",
		EnableBashCompletion: true,
		Usage:                "Find un used packages fast",
		Flags:                flags,
		// Before:               altsrc.InitInputSourceWithContext(flags, ),
		Action: func(c *cli.Context) error {
			loadConfigFromFile()

			depcheck()

			return nil
		},
		Commands: commands,
	}

	return *app

}
