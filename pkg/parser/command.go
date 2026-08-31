package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func returnToFileScope() {
	fsIndex := len(fileScopes) - 1
	if fsIndex >= 0 {
		scope = fileScopes[fsIndex]
	} else {
		scope = 0
	}
	descopeVars()
}

func enterScope() {
	scope++
}

func exitScope() {
	scope--
	fsIndex := len(fileScopes) - 1
	for fsIndex >= 0 && fileScopes[fsIndex] > scope {
		fileScopes = fileScopes[0:fsIndex]
		fsIndex--
	}
	descopeVars()
}

func enterFileScope() {
	enterScope()
	fileScopes = append(fileScopes, scope)
}

func assignManifest() {
	putVarOnTopHookless("_songsManifest", ":songs")
	songManifests := variables["_songsManifest"]
	currentManifest = (*songManifests)[len(*songManifests)-1]
}

func descopeVars() {
	deletableVars := []string{}
	for varName, varEntry := range variables {
		entryIndex := len(*varEntry) - 1
		for entryIndex >= 0 && (*varEntry)[entryIndex].scope > scope {
			*varEntry = (*varEntry)[0:entryIndex]
			entryIndex--
		}
		if entryIndex == -1 {
			deletableVars = append(deletableVars, varName)
		}
	}
	for _, varName := range deletableVars {
		delete(variables, varName)
		delete(variableHooks, varName)
	}
	for _, varEntry := range variableHooks {
		deletableHooks := []string{}
		for hookName, hookEntry := range *varEntry {
			entryIndex := len(*hookEntry) - 1
			for entryIndex >= 0 && (*hookEntry)[entryIndex].scope > scope {
				*hookEntry = (*hookEntry)[0:entryIndex]
				entryIndex--
			}
			if entryIndex == -1 {
				deletableHooks = append(deletableHooks, hookName)
			}
		}
		for _, hookName := range deletableHooks {
			delete(*varEntry, hookName)
		}
	}
}

func exploreUsing() {
	using, _ := getVarOnTopHookless("using")
	manifestFilePath, _ := getVarOnTopHookless("_srcStack")
	manifestFolderPath := filepath.Dir(manifestFilePath)
	destFolderPath, _ := getVarOnTopHookless("_destStack")

	paths := []string{}
	if using == "*" {
		files, _ := os.ReadDir(manifestFolderPath)
		for _, file := range files {
			if file.IsDir() {
				paths = append(paths, filepath.Join(manifestFolderPath, file.Name(), "manifest.txt"))
			}
		}
	} else {
		folderNames := strings.SplitSeq(using, "&&")
		for rawName := range folderNames {
			trimmedName := strings.TrimSpace(rawName)
			if len(trimmedName) > 0 {
				paths = append(paths, filepath.Join(manifestFolderPath, trimmedName, "manifest.txt"))
			}
		}
	}
	heldManifest := currentManifest
	heldManifest.value = fmt.Sprintf("%s\n:folders", heldManifest.value)
	for _, path := range paths {
		folderBase := filepath.Base(filepath.Dir(path))
		if _, err := os.Stat(path); err == nil {
			if err := ParseAndExecute(path, filepath.Join(destFolderPath, folderBase)); err != nil {
				fmt.Printf("Could not parse %s\n\tCause: %s\n", path, err)
			} else {
				heldManifest.value = fmt.Sprintf("%s\n\t%s", heldManifest.value, folderBase)
			}
		}
	}
	currentManifest = heldManifest
}

func writeManifest() {
	destFolderPath, _ := getVarOnTopHookless("_destStack")
	destPath := filepath.Join(destFolderPath, "manifest.txt")
	content := currentManifest.value
	if content != ":songs" {
		w, err := os.Create(destPath)
		if err != nil {
			// unhandled for now
			return
		}
		defer w.Close()
		fmt.Fprint(w, content)
	}
}
