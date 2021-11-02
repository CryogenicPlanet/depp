package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/romanyx/mdopen"
)

var report bool
var showVersions bool

func checkIgnoredNameSpace(module string) bool {
	namespaces := ignoreNameSpaces.Value()

	if len(namespaces) > 0 {
		for _, name := range namespaces {
			if strings.Contains(module, name) {
				return true
			}
		}
	}
	return false
}

func computeResults() {

	count := 0
	unused := 0
	unusedPackages := []string{}
	yellow := color.New(color.FgYellow)
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	for key, val := range deps {

		if checkIgnoredNameSpace(key) {
			continue
		}

		if checkTypePackage(key) {
			continue
		}

		if !val {
			yellow.Println("Package", key, "is unused")
			fileLog("Package", key, "is unused")
			unusedPackages = append(unusedPackages, key)
			unused += 1
		} else {
			green.Println("Package", key, "is used!")
			fileLog("Package", key, "is used!")
		}
		count += 1
	}

	fmt.Print("There are a total of ", count, " packages and ")
	yellow.Println(unused, "are the unused")

	fileLog("There are a total of ", count, " packages and ", unused, " are the unused")

	reportLog("**There are a total of ", count, " packages and ", unused, " are the unused**")

	fmt.Print("The unused packages are ")

	yellow.Println(strings.Join(unusedPackages, ", "))

	unusedPackageMarkdownTable(unusedPackages)

	duplicatesDiffVersions, duplicatesSameVersion := findDuplicates()

	unusedTypePackage := checkAtTypesPackages()

	if len(duplicatesSameVersion) > 0 || len(duplicatesDiffVersions) > 0 {
		if len(duplicatesSameVersion) > 0 {

			fmt.Print("The duplicate packages with same versions are ")
			yellow.Println(strings.Join(duplicatesSameVersion, ", "))
		}
		if len(duplicatesDiffVersions) > 0 {

			fmt.Print("The duplicate packages with different versions are ")
			red.Println(strings.Join(duplicatesDiffVersions, ", "))
		}

	} else {
		green.Println("There are no duplicate packages")
	}

	unusedTypesPackagesMarkdownTable(unusedTypePackage)

}

func findDuplicates() ([]string, []string) {
	duplicatesDiffVersionName := []string{}
	duplicatesDiffVersions := [][]string{}
	duplicatesSameVersion := []string{}

	yellow := color.New(color.FgYellow)

	for key, val := range versions {

		if !checkIgnoredNameSpace(key) {
			uniqueVersions := uniqueStringSlice(val)
			if len(val) > 1 {

				if len(uniqueVersions) > 1 {
					duplicatesDiffVersionName = append(duplicatesDiffVersionName, key)
					duplicatesDiffVersions = append(duplicatesDiffVersions, uniqueVersions)
				} else {
					duplicatesSameVersion = append(duplicatesSameVersion, key)
				}
				if showVersions {
					if len(uniqueVersions) > 1 {
						fmt.Print("The package ")
						yellow.Print(key)
						fmt.Print(" has multiple versions - ")
						yellow.Println(strings.Join(uniqueVersions, ", "))
					}
				}
			}
		}
	}
	reportLog("## Duplicate packages")

	sameVersionMarkdownTable(duplicatesSameVersion)

	multipleVersionsMarkdownTable(duplicatesDiffVersionName, duplicatesDiffVersions)

	return duplicatesDiffVersionName, duplicatesSameVersion
}

var reportLines = make(chan string, 100)

var reportWg sync.WaitGroup

// Log to generate report file
func reportLog(a ...interface{}) {
	str := fmt.Sprint(a...)
	reportWg.Add(1)
	reportLines <- str

}

func unusedPackageMarkdownTable(packages []string) {
	if len(packages) > 0 {
		reportLog("## Unused packages \n")
		for _, val := range packages {
			reportLog("- [] ", val)
		}
		reportLog("\n---")
	}
}

func multipleVersionsMarkdownTable(packages []string, packageVersions [][]string) {
	if len(packages) > 0 {
		reportLog("### Packages with Multiple Versions")
		reportLog("| Package  | Version | Used By")
		reportLog("| ----------- | ----------- | ----------- |")
		for i := range packages {
			name := packages[i]
			versions := packageVersions[i]
			reportLog("| ", name, " | `", strings.Join(versions, ","), "` | `", strings.Join(depsName[name], ", "), "` | ")
		}
		reportLog("---")
	}
}

func sameVersionMarkdownTable(packages []string) {
	if len(packages) > 0 {
		reportLog("### Packages with Same Versions")
		reportLog("| Package  | Used By |")
		reportLog("| ----------- | ----------- |")
		for _, val := range packages {
			reportLog("| ", val, " | `", strings.Join(depsName[val], ", "), "` | ")
		}
		reportLog("---")
	}
}

func generateReport() {

	// open output file
	fo, err := os.Create(DEPCHECK_DIR + "/report.md")
	fileOps.Add(1)
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error

	datawriter := bufio.NewWriter(fo)

	for line := range reportLines {
		// fmt.Println("Writing", line, "to report")
		_, err := datawriter.WriteString(line + "\n")
		check(err)
		reportWg.Done()
	}

	err = datawriter.Flush()
	check(err)
	err = fo.Close()
	check(err)
	fileOps.Done()

}

func openHtml() {
	if !noOpen {
		md, err := ioutil.ReadFile(DEPCHECK_DIR + "/report.md")

		if err != nil {
			fmt.Println("Previous report not found, please generate a report")
			os.Exit(1)
		}

		reader := bytes.NewReader(md)

		// err := os.WriteFile(DEPCHECK_DIR+"/report.html", output, 0644)

		// if err != nil {
		// 	panic(err)
		// }

		opnr := mdopen.New()
		if err := opnr.Open(reader); err != nil {
			log.Fatal(err)
		}
		// err = browser.OpenURL(DEPCHECK_DIR + "/report.html")

		// if err != nil {
		// 	panic(err)
		// }
	}

}
