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
)

type PackageJSON struct {
	Name            string            `json:"name"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
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

	if globalConfig.JS {

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

			if globalConfig.DevDependencies {
				deps[key] = false
			} else {
				if checkTypePackage(key) {
					deps[key] = false
				}
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

	if globalConfig.Path != "" {
		rootPath = globalConfig.Path
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

	platform := api.PlatformNode

	if globalConfig.BrowserPlatform {
		platform = api.PlatformBrowser
	}

	result := api.Build(api.BuildOptions{
		EntryPoints: sourcePaths,
		// EntryPoints: []string{"test/monorepo/packages/package-b/src/App.tsx"},
		Target:   api.ESNext,
		Bundle:   true,
		Write:    esbuildWrite,
		Format:   api.FormatESModule,
		Outdir:   DEPCHECK_DIR + "/dist",
		Plugins:  []api.Plugin{depcheckPlugin},
		External: globalConfig.Externals,
		Metafile: true,
		Platform: platform,
		Loader: map[string]api.Loader{
			".js": api.LoaderJSX,
		},
	})

	if len(result.Errors) > 0 {

		for _, err := range result.Errors {
			fmt.Println("Error", err.Text, err.Location.File, err.Location.Line)
			fileLog("Error", err.Text, err.Location.File, err.Location.Line, err)
		}

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

	deployUrl := ""

	if netlifyToken != "" {
		deployUrl = deployToNetlify()
		fmt.Print("Deployed URL: ")
		color.New(color.FgGreen).Println(deployUrl)
	}

	if ciMode {
		makePrComment(deployUrl)
	}

	// Auto deletes the folder by default
	// The folder is used to create the html report everytime
	if !globalConfig.Log && !globalConfig.Report && !esbuildWrite && !hasConfig && !saveConfig {

		removeDirectory(true)
	}

	if saveConfig {
		writeConfig(globalConfig)
	}

	close(modules)
}

func main() {
	netlifyPat := os.Getenv("NETLIFY_TOKEN")
	if netlifyToken == "" {
		netlifyToken = netlifyPat
	}

	githubPat := os.Getenv("GITHUB_TOKEN")

	githubToken = githubPat

	app := createCliApp()

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
