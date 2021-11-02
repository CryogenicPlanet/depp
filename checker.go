package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/tidwall/gjson"
)

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

// Check @types packages in the packagejson
func checkAtTypesPackages() []string {

	unused := []string{}

	yellow := color.New(color.FgYellow)

	for packageName, _ := range deps {
		if checkTypePackage(packageName) {
			// This is a @type package
			originalName := strings.Split(packageName, "/")[1]
			if _, ok := deps[originalName]; !ok {
				unused = append(unused, packageName)
			}

		}
	}
	fmt.Print("The unused '@types' packages are ")
	yellow.Println(strings.Join(unused, ", "))

	return unused

}

func unusedTypesPackagesMarkdownTable(unusedTypes []string) {
	if len(unusedTypes) > 0 {
		reportLog("## Unused `@types` packages")
		reportLog("| Type Package  | Actual Package | Used By |")
		reportLog("| ----------- | ----------- | ----------- |")

		for _, val := range unusedTypes {
			originalName := strings.Split(val, "/")[1]
			reportLog("| ", val, " | ", originalName, " | `", strings.Join(depsName[val], ", "), "` | ")
		}
		reportLog("---")
	}
}

func checkTypePackage(name string) bool {
	return strings.Contains(name, "@types/")
}
