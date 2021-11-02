package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"log"
	"os"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/fatih/color"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v2"
)

type PackageJSON struct {
	Name            string            `json: "name, omitempty"`
	Dependencies    map[string]string `json: "dependencies, omitempty"`
	DevDependencies map[string]string `json "devDependencies, omitempty"`
}

var depcheckPlugin = api.Plugin{
	Name: "Depcheck Plugin",
	Setup: func(build api.PluginBuild) {

		// Everything
		build.OnResolve(api.OnResolveOptions{Filter: `\.`},
			func(args api.OnResolveArgs) (api.OnResolveResult, error) {

				if strings.Contains(args.Importer, "node_module") {

					return api.OnResolveResult{
						Path:     args.Path,
						External: true,
					}, nil
				}

				ext := filepath.Ext(args.Path)

				if strings.Contains(ext, "ts") || strings.Contains(ext, "js") || ext == "" {

					// if strings.Contains(args.Importer, "node_modules") {
					// 	// Ignore node modules
					// 	return api.OnResolveResult{
					// 		Path:     args.Path,
					// 		External: true,
					// 	}, nil
					// }

					// fmt.Println("isTsJsLike")
					return api.OnResolveResult{}, nil

				}

				return api.OnResolveResult{
					Path:     args.Path,
					External: true,
				}, nil
			})

		build.OnResolve(api.OnResolveOptions{Filter: `[t|j]sx?$`},
			func(args api.OnResolveArgs) (api.OnResolveResult, error) {

				fileLog("Resolved", args, args.Kind)

				// if strings.Contains(args.Importer, "node_modules") {

				// 	importer := args.Importer
				// 	lastNM := strings.LastIndex(importer, "node_module")

				// 	if lastNM == -1 {
				// 		color.New(color.FgYellow).Println("[WARN] Error finding node_module in require call, skipping", importer)
				// 		return api.OnResolveResult{}, nil
				// 	}

				// 	str := importer[lastNM:]

				// 	splitBySlash := strings.Split(str, "/")
				// 	// fmt.Println(importer, lastNM, str, splitBySlash)
				// 	moduleName := splitBySlash[1]

				// 	if strings.Contains(moduleName, "@") {
				// 		// using a @x/y package
				// 		// Example @babel/core
				// 		moduleName += "/" + splitBySlash[2]
				// 	}

				// 	checkModule(moduleName)

				// 	// fmt.Println("Skipping", args.Importer)
				// 	fileLog("Skipping", args)
				// 	return api.OnResolveResult{
				// 		External: true,
				// 	}, nil
				// }

				if args.Kind == api.ResolveJSImportStatement {
					// import statement
					path := args.Path
					fileLog("Import statement", args)

					importer := args.Importer
					lastNM := strings.LastIndex(importer, "node_module")

					if lastNM == -1 {
						// color.New(color.FgYellow).Println("[WARN] Error finding node_module in require call, skipping", importer)
						return api.OnResolveResult{}, nil
					}

					str := importer[lastNM:]

					splitBySlash := strings.Split(str, "/")
					// fmt.Println(importer, lastNM, str, splitBySlash)
					moduleName := splitBySlash[1]

					if strings.Contains(moduleName, "@") {
						// using a @x/y package
						// Example @babel/core
						moduleName += "/" + splitBySlash[2]
					}

					fileLog("Import statement", path, moduleName, args.Kind)
					checkModule(path)

					checkModule(moduleName)

					return api.OnResolveResult{
						Path:     args.Path,
						External: true,
					}, nil
				}

				if args.Kind == api.ResolveJSRequireCall {
					// require call

					importer := args.Importer
					lastNM := strings.LastIndex(importer, "node_module")
					fileLog("Require call", args)

					if lastNM == -1 {
						// color.New(color.FgYellow).Println("[WARN] Error finding node_module in require call, skipping", importer)
						return api.OnResolveResult{}, nil
					}

					str := importer[lastNM:]

					splitBySlash := strings.Split(str, "/")
					// fmt.Println(importer, lastNM, str, splitBySlash)
					moduleName := splitBySlash[1]

					if strings.Contains(moduleName, "@") {
						// using a @x/y package
						// Example @babel/core
						moduleName += "/" + splitBySlash[2]
					}

					fileLog("Require call", moduleName)

					checkModule(moduleName)

					return api.OnResolveResult{
						Path:     args.Path,
						External: true,
					}, nil

				}

				// return api.OnResolveResult{
				// 	Path: filepath.Join(args.ResolveDir, "public", args.Path),
				// }, nil

				return api.OnResolveResult{}, nil

			})

		build.OnLoad(api.OnLoadOptions{Filter: `\.`},
			func(args api.OnLoadArgs) (api.OnLoadResult, error) {
				path := args.Path

				ext := filepath.Ext(path)

				if ext != "" {
					// has file extension
					if strings.Contains(ext, "ts") || strings.Contains(ext, "js") {
						return api.OnLoadResult{}, nil
					}
					// default to file loader for all other file types
					return api.OnLoadResult{Loader: api.LoaderFile}, nil
				}
				return api.OnLoadResult{}, nil
			})
	},
}

func getPackageJsonPaths(path string) []string {
	files, err := Glob(path + "**/package.json")

	if err != nil {
		fmt.Println("Error Reading file paths from glob", err)
	}
	// fmt.Println("Package json files are", files)
	return files
}

func getFiles(path string) []string {
	var files []string

	tsFiles, err := Glob(path + "**/*.ts*")

	for _, file := range tsFiles {

		ext := filepath.Ext(file)

		// The ext check prevents .ts.snap or ts.anything files\
		// By default we will check .test.ts files
		// TODO Add a flag for ignored directories later
		if !strings.Contains(file, ".d.ts") && (ext == ".ts" || ext == ".tsx") {
			files = append(files, file)
		}
	}

	if err != nil {
		fmt.Println("Error Reading file paths from glob", err)
	}

	if jsSource {

		jsFiles, err := Glob(path + "**/*.js")

		if err != nil {
			fmt.Println("Error Reading file paths from glob", err)
		}
		files = append(files, jsFiles...)

		// js* === json and hence we need to have a separate case for *.jsx
		jsxFiles, err := Glob(path + "**/*.jsx")

		if err != nil {
			fmt.Println("Error Reading file paths from glob", err)
		}
		files = append(files, jsxFiles...)

	}

	// fmt.Println("Source Files are", files)
	return files
}

func readJson(path string) PackageJSON {

	plan, _ := ioutil.ReadFile(path)

	packageJson := PackageJSON{}
	err := json.Unmarshal(plan, &packageJson)

	if err != nil {
		fmt.Println("Error reading file", err)
		os.Exit(1)
	}
	return packageJson
}

var deps = make(map[string]bool)
var depsName = make(map[string][]string)
var versions = make(map[string][]string)

func setDeps(paths []string) {
	for _, path := range paths {
		packageJson := readJson(path)

		for key, version := range packageJson.Dependencies {
			deps[key] = false

			if _, ok := depsName[key]; ok {
				depsName[key] = append(depsName[key], packageJson.Name)
			} else {
				depsName[key] = []string{packageJson.Name}
			}

			if _, ok := versions[key]; ok {
				versions[key] = append(versions[key], version)
			} else {
				versions[key] = []string{version}
			}

		}

		for key, version := range packageJson.DevDependencies {
			if devDep {
				deps[key] = false
			}
			if _, ok := versions[key]; ok {
				versions[key] = append(versions[key], version)
			} else {
				versions[key] = []string{version}
			}

			if _, ok := depsName[key]; ok {
				depsName[key] = append(depsName[key], packageJson.Name)
			} else {
				depsName[key] = []string{packageJson.Name}
			}
		}

	}
}

// This function will go through the MetaFile to double check any missed imports
func checkMetaFile(metaFile string, rootPath string, sourcePaths []string) bool {

	metafile := gjson.Parse(metaFile).Value().(map[string]interface{})

	inputs := metafile["inputs"].(map[string]interface{})
	// fmt.Println("Inputs", inputs)
	newInputs := map[string]interface{}{}
	folders := strings.Split(rootPath, "/")
	lastFolder := folders[len(folders)-1]

	for key, val := range inputs {
		if strings.Contains(key, "node_modules") {
			continue
		}
		// Imported files
		// fmt.Println("Updating key", key, "using last folder", lastFolder)

		if index := strings.Index(key, lastFolder); index != -1 {
			newKey := key[index:]
			newInputs[newKey] = val
		} else {
			newKey := lastFolder + "/" + key
			newInputs[newKey] = val
		}

	}

	for _, source := range sourcePaths {

		index := strings.Index(source, lastFolder)

		trimmedSource := source[index:]

		if inputFileInter, ok := newInputs[trimmedSource]; ok {

			inputFile := inputFileInter.(map[string]interface{})

			for _, importCallInt := range inputFile["imports"].([]interface{}) {
				importCall := importCallInt.(map[string]interface{})

				// fmt.Println("Handling import call", importCall, "from source", trimmedSource)
				if importCall["kind"] == "import-statement" || importCall["kind"] == "require-call" {
					path := importCall["path"].(string)
					lastNM := strings.LastIndex(path, "node_module")

					if lastNM == -1 {
						// color.New(color.FgYellow).Println("[WARN] Error finding node_module in require call, skipping", path)
						break
					}

					str := path[lastNM:]

					splitBySlash := strings.Split(str, "/")
					// fmt.Println(importer, lastNM, str, splitBySlash)
					moduleName := splitBySlash[1]

					if strings.Contains(moduleName, "@") {
						// using a @x/y package
						// Example @babel/core
						moduleName += "/" + splitBySlash[2]
					}

					checkModule(moduleName)

				}
			}
		}
	}
	return true
}

var overwriteRootPath string

var overwriteSource string

func depcheck() {
	createDirectory()

	if esbuildWrite {
		color.New(color.FgRed).Println("Do not use this as a bundler of files, it will likely not work")
		fmt.Println("This write mode is really only for debugging purposes")
	}

	go writeLogsToFile()
	go handleModule()
	go generateReport()

	rootPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	if overwriteRootPath != "" {
		rootPath = overwriteRootPath
	}

	fmt.Println("Root path:", rootPath)

	packageJsonPaths := getPackageJsonPaths(rootPath)

	sourcePaths := getFiles(rootPath)

	if overwriteSource != "" {
		sourcePaths = []string{overwriteSource}
		fmt.Println("Overwritten source", sourcePaths)
	}

	for _, path := range sourcePaths {
		fileLog("Source Path", path)
	}

	setDeps(packageJsonPaths)

	result := api.Build(api.BuildOptions{
		EntryPoints: sourcePaths,
		// EntryPoints: []string{"test/monorepo/packages/package-b/src/App.tsx"},
		Target:   api.ES2016,
		Bundle:   true,
		Write:    esbuildWrite,
		Format:   api.FormatESModule,
		Outdir:   DEPCHECK_DIR + "/dist",
		Plugins:  []api.Plugin{depcheckPlugin},
		External: getExternals(),
		Metafile: true,
	})

	if len(result.Errors) > 0 {
		fmt.Println("Errors", result.Errors)

		os.Exit(1)
	}
	checkMetaFile(result.Metafile, rootPath, sourcePaths)

	moduleWg.Wait()

	reportLog("# Report for - ", rootPath)

	computeResults()

	// reportWg.Add(1) // Extra wg to allow file to finish writing
	reportWg.Wait()
	close(reportLines)

	loggerWg.Wait()
	close(logs)

	fileOps.Wait()

	openHtml()

	close(modules)
}

var devDep bool
var jsSource bool
var esbuildWrite bool

var externals cli.StringSlice
var ignoreNameSpaces cli.StringSlice

func main() {
	app := &cli.App{
		Name:  "depp",
		Usage: "Find un used packages fast",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "dev",
				Aliases:     []string{"d"},
				Usage:       "Enable dev dependencies",
				Destination: &devDep,
			},
			&cli.BoolFlag{
				Name:        "js",
				Aliases:     []string{"j"},
				Usage:       "Enable js source files",
				Destination: &jsSource,
			},
			&cli.PathFlag{
				Name:        "path",
				Aliases:     []string{"p"},
				Usage:       "Overwrite root directory",
				Destination: &overwriteRootPath,
			},
			&cli.BoolFlag{
				Name:        "log",
				Aliases:     []string{"l"},
				Usage:       "Will write logs to .depcheck.log",
				Value:       false,
				Destination: &logging,
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
				Destination: &report,
			},
			&cli.BoolFlag{
				Name:        "show-versions",
				Aliases:     []string{"v"},
				Usage:       "Show conflicting versions",
				Value:       false,
				Destination: &showVersions,
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
				Name:        "ignore-namespace",
				Aliases:     []string{"in"},
				Usage:       "Pass namespace (@monorepo) to be ignored",
				Destination: &ignoreNameSpaces,
			},
		},
		Action: func(c *cli.Context) error {
			depcheck()

			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "clean",
				Usage: "Cleans all output files",
				Action: func(c *cli.Context) error {
					removeDirectory()
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
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
