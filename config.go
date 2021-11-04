package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli/v2"
)

func getExternals() []string {
	defaultExternals := []string{"path", "fs", "crypto", "os", "http", "child_process", "querystring", "readline", "tls", "assert", "buffer", "url", "net", "buffer", "tty", "util", "stream", "events", "zlib", "https", "worker_threads", "module", "http2", "dns"}

	cliExternals := externals.Value()

	if len(cliExternals) > 0 {
		fmt.Println("Cli externals", externals.Value())
		defaultExternals = append(defaultExternals, cliExternals...)
	}

	return defaultExternals
}

type ConfigJSON struct {
	JS                bool     `json:"js"`
	Path              string   `json:"path"`
	Log               bool     `json:"log"`
	Report            bool     `json:"report"`
	Versions          bool     `json:"show-versions"`
	DevDependencies   bool     `json:"dev"`
	Externals         []string `json:"externals"`
	IgnoredNamespaces []string `json:"ignore-namespaces"`
}

func setupConfig() {
	createDirectory()

	config := ConfigJSON{JS: false, Report: true}

	prompt := &survey.Confirm{
		Message: "Enable javascript (Default typescript only)",
	}
	survey.AskOne(prompt, &config.JS)

	prompt = &survey.Confirm{
		Message: "Show versions of duplicate packages",
	}
	survey.AskOne(prompt, &config.Versions)

	prompt = &survey.Confirm{
		Message: "Check dev dependencies (unstable)",
	}
	survey.AskOne(prompt, &config.DevDependencies)

	writeConfig((config))
}

func setVersions() {
	config := retriveConfig()

	prompt := &survey.Confirm{
		Message: "Show versions of duplicate packages",
	}
	survey.AskOne(prompt, &config.Versions)

	writeConfig(config)
}

func setJs() {
	config := retriveConfig()

	prompt := &survey.Confirm{
		Message: "Enable javascript (Default typescript only)",
	}
	survey.AskOne(prompt, &config.JS)

	writeConfig(config)
}

func writeConfig(config ConfigJSON) {
	configJson, _ := json.Marshal(config)
	err := ioutil.WriteFile(DEPCHECK_DIR+"/config.json", configJson, 0644)

	if err != nil {
		panic(err)
	}
}

func retriveConfig() ConfigJSON {

	configJson, err := ioutil.ReadFile(DEPCHECK_DIR + "/config.json")

	if err != nil {
		fmt.Println("No config found, run depp init")
		os.Exit(1)
	}

	config := ConfigJSON{}

	err = json.Unmarshal(configJson, &config)

	if err != nil {
		fmt.Println("Not intialized config properly")
		os.Exit(1)
	}

	return config
}

func showConfig() {
	config := retriveConfig()

	fmt.Println("The current config is")

	v := reflect.ValueOf(config)
	typeOfS := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fmt.Printf("%s\tValue: %v\n", typeOfS.Field(i).Name, v.Field(i).Interface())
	}

}

func setPath(input string) {
	config := retriveConfig()

	if input == "" {
		prompt := &survey.Input{
			Message: "Set Root Path",
		}
		survey.AskOne(prompt, &input)
	}
	config.Path = input

	writeConfig(config)
}

func setExternals(lines []string) {
	config := retriveConfig()

	if len(lines) == 0 {

		input := ""

		prompt := &survey.Multiline{
			Message: "Set Externals (each line is one package)",
		}
		survey.AskOne(prompt, &input)

		lines = strings.Split(input, "\n")
	}

	config.Externals = lines

	writeConfig(config)
}

func addExternal(input string) {
	config := retriveConfig()

	if input == "" {
		prompt := &survey.Input{
			Message: "Package to add as external",
		}
		survey.AskOne(prompt, &input)
	}

	config.Externals = append(config.Externals, input)

	writeConfig(config)
}

func removeExternal(input string) {
	config := retriveConfig()

	if input == "" {

		prompt := &survey.Input{
			Message: "Package to remove as external",
		}
		survey.AskOne(prompt, &input)
	}

	index := -1

	for i, val := range config.Externals {
		if val == input {
			index = i
		}
	}

	config.Externals = append(config.Externals[:index], config.Externals[index+1:]...)

	writeConfig(config)
}

func setINS(lines []string) {
	config := retriveConfig()

	if len(lines) == 0 {
		input := ""

		prompt := &survey.Multiline{
			Message: "Set Externals (each line is one package)",
		}
		survey.AskOne(prompt, &input)

		lines = strings.Split(input, "\n")
	}

	config.IgnoredNamespaces = lines

	writeConfig(config)
}

func addINS(input string) {
	config := retriveConfig()

	if input == "" {
		prompt := &survey.Input{
			Message: "Package to add as external",
		}
		survey.AskOne(prompt, &input)
	}

	config.IgnoredNamespaces = append(config.IgnoredNamespaces, input)

	writeConfig(config)
}

func removeINS(input string) {
	config := retriveConfig()

	if input == "" {

		prompt := &survey.Input{
			Message: "Package to remove as external",
		}
		survey.AskOne(prompt, &input)
	}

	index := -1

	for i, val := range config.IgnoredNamespaces {
		if val == input {
			index = i
		}
	}

	config.IgnoredNamespaces = append(config.IgnoredNamespaces[:index], config.IgnoredNamespaces[index+1:]...)

	writeConfig(config)
}

func setupConfigCLI() []*cli.Command {

	commands := []*cli.Command{
		{
			Name:  "config",
			Usage: "A command to handle config",
			Subcommands: []*cli.Command{
				{
					Name:  "show",
					Usage: "Show current configuration",
					Action: func(c *cli.Context) error {
						showConfig()
						return nil
					},
				},
				{
					Name:  "path",
					Usage: "Set the root path",
					Action: func(c *cli.Context) error {
						setPath(c.Args().Get(0))
						return nil
					},
				},
				{
					Name:  "versions",
					Usage: "Set versions config",
					Action: func(c *cli.Context) error {
						setVersions()
						return nil
					},
				},
				{
					Name:  "js",
					Usage: "Set js config",
					Action: func(c *cli.Context) error {
						setJs()
						return nil
					},
				},
				{
					Name:  "externals",
					Usage: "Handle external config",
					Subcommands: []*cli.Command{
						{
							Name:  "add",
							Usage: "Add external",
							Action: func(c *cli.Context) error {
								addExternal(c.Args().Get(0))
								return nil
							},
						},
						{
							Name:  "set",
							Usage: "Set externals",
							Action: func(c *cli.Context) error {
								args := c.Args().Tail()
								args = append(args, c.Args().First())
								setExternals(args)
								return nil
							},
						},
						{
							Name:  "remove",
							Usage: "Remove external",
							Action: func(c *cli.Context) error {
								removeExternal(c.Args().Get(0))
								return nil
							},
						},
					},
				},
				{
					Name:  "ignored-namespaces",
					Usage: "Handle ignored namespace config",
					Subcommands: []*cli.Command{
						{
							Name:  "add",
							Usage: "Add ignored namespace",
							Action: func(c *cli.Context) error {
								addINS(c.Args().Get(0))
								return nil
							},
						},
						{
							Name:  "set",
							Usage: "Set ignored namespaces",
							Action: func(c *cli.Context) error {
								args := c.Args().Tail()
								args = append(args, c.Args().First())
								setINS(args)
								return nil
							},
						},
						{
							Name:  "remove",
							Usage: "Remove ignored namespace",
							Action: func(c *cli.Context) error {
								removeINS(c.Args().Get(0))
								return nil
							},
						},
					},
				},
			},
		},
		{
			Name:  "init",
			Usage: "Initialize project",
			Action: func(c *cli.Context) error {
				fmt.Println("Init project")
				setupConfig()
				return nil
			},
		},
	}

	return commands

}
