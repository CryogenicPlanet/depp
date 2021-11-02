package main

import (
	"os"
	"path/filepath"
	"strings"
)

// Expand finds matches for the provided Globs.
func handleGlobs(globs []string, ignored []string) ([]string, error) {
	var matches = []string{""} // accumulate here

	for _, glob := range globs {
		if glob != "" && string(glob[0]) == "." {
			// Check if first character is "."
			// This eliminates cases where you have .ts instead of *.ts
			glob = "*" + glob
		}
		if glob == "" {
			glob = "."
		}

		// fmt.Println("Glob iteration", glob)
		var hits []string
		var hitMap = map[string]bool{}
		for _, match := range matches {
			// fmt.Println("Match iteration", match)
			paths, err := filepath.Glob(match + glob)
			// fmt.Println("Paths", paths)
			if err != nil {
				return nil, err
			}
			for _, path := range paths {
				err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					// save de duped match from current iteration
					if _, ok := hitMap[path]; !ok {

						for _, ignorePattern := range ignored {
							ok, err := filepath.Match(ignorePattern, path)
							if ok || err != nil || strings.Contains(path, ignorePattern) {
								// is an ignored file
								return nil
							}

						}

						if strings.Contains(path, "node_modules") {
							return nil
						}

						hits = append(hits, path)
						hitMap[path] = true
					}
					return nil
				})
				if err != nil {
					return nil, err
				}
			}
		}
		matches = hits
	}

	// fix up return value for nil input
	if globs == nil && len(matches) > 0 && matches[0] == "" {
		matches = matches[1:]
	}

	return matches, nil
}

// Globs represents one filepath glob, with its elements joined by "**".

// Glob adds double-star support to the core path/filepath Glob function.
// It's useful when your globs might have double-stars, but you're not sure.
func Glob(pattern string) ([]string, error) {
	if !strings.Contains(pattern, "**") {
		// passthru to core package if no double-star
		return filepath.Glob(pattern)
	}
	ingored := []string{"node_modules", "dist", ".next"}

	return handleGlobs(strings.Split(pattern, "**"), ingored)
}

func uniqueStringSlice(intSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
