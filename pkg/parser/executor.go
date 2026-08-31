package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/frogssoldseparately/randoMusicManager/pkg/mmrs"
)

var variables map[string]*[]*variableValue
var variableHooks map[string]*map[string]*[]*hookValue
var bio map[string]string
var scope int
var fileScopes []int
var currentManifest *variableValue

type variableValue struct {
	value string
	scope int
}

type hookValue struct {
	args  []string
	scope int
}

func Setup() {
	variables = map[string]*[]*variableValue{}
	variables["filestart"] = &[]*variableValue{{"", 0}}
	variables["fileend"] = &[]*variableValue{{"", 0}}
	variables["_srcStack"] = &[]*variableValue{{"", 0}}
	variables["_destStack"] = &[]*variableValue{{"", 0}}
	variables["_songsManifest"] = &[]*variableValue{{":songs", 0}}
	variableHooks = map[string]*map[string]*[]*hookValue{}
	scope = 0
	fileScopes = []int{0}
	bio = map[string]string{
		"title":     "",
		"composers": "",
		"seq":       "",
		"midi":      "",
	}
	currentManifest = (*variables["_songsManifest"])[0]
}

func ParseAndExecute(srcManifest string, destFolder string) error {
	p := MakeParser()
	statements, err := p.ParseFile(srcManifest)
	if err != nil {
		return err
	}
	encloseStatements(statements, srcManifest, destFolder)
	return execute(*statements, srcManifest, destFolder)
}

func encloseStatements(statements *[]*Statement, src string, dest string) {
	firstStatements := []*Statement{
		{assignmentId, []*Expression{
			{varSetExpId, []*Token{
				{varSetId, "filestart", 0, 0},
			}, 0, 0},
			{stringlitId, []*Token{
				{stringId, "", 0, 0},
			}, 0, 0},
		}, 0, 0},
		{assignmentId, []*Expression{
			{varSetExpId, []*Token{
				{varSetId, "_srcStack", 0, 0},
			}, 0, 0},
			{stringlitId, []*Token{
				{stringId, src, 0, 0},
			}, 0, 0},
		}, 0, 0},
		{assignmentId, []*Expression{
			{varSetExpId, []*Token{
				{varSetId, "_destStack", 0, 0},
			}, 0, 0},
			{stringlitId, []*Token{
				{stringId, dest, 0, 0},
			}, 0, 0},
		}, 0, 0},
	}
	lastStatements := []*Statement{
		{assignmentId, []*Expression{
			{varSetExpId, []*Token{
				{varSetId, "fileend", 0, 0},
			}, 0, 0},
			{stringlitId, []*Token{
				{stringId, "", 0, 0},
			}, 0, 0},
		}, 0, 0},
	}
	*statements = slices.Concat(firstStatements, *statements, lastStatements)
}

func getVarOnTopHookless(varName string) (string, error) {
	if stack, ok := variables[varName]; ok {
		return (*stack)[len(*stack)-1].value, nil
	}
	return "", fmt.Errorf("No such variable \"%s\"", varName)
}

func getVarOnTop(varName string) (string, error) {
	if err := runHooks(varName, "prehookget"); err != nil {
		return "", err
	}
	out, err := getVarOnTopHookless(varName)
	if err != nil {
		return "", err
	}
	if err := runHooks(varName, "posthookget"); err != nil {
		return "", err
	}
	return out, nil
}

func putVarOnTopHookless(varName string, value string) {
	if stack, ok := variables[varName]; ok {
		lastIndex := len(*stack) - 1
		if (*stack)[lastIndex].scope == scope {
			(*stack)[lastIndex].value = value
		} else {
			*stack = append(*stack, &variableValue{value, scope})
		}
	} else {
		variables[varName] = &[]*variableValue{{value, scope}}
	}
}

func putVarOnTop(varName string, value string) error {
	if err := runHooks(varName, "prehookset"); err != nil {
		return err
	}
	putVarOnTopHookless(varName, value)
	return runHooks(varName, "posthookset")
}

func putHookOnTop(varName string, funcName string, args []string) {
	varGroup, ok := variableHooks[varName]
	if !ok {
		varGroup = &map[string]*[]*hookValue{}
		variableHooks[varName] = varGroup
	}
	funcGroup, ok := (*varGroup)[funcName]
	if !ok {
		funcGroup = &[]*hookValue{}
		(*varGroup)[funcName] = funcGroup
	}
	*funcGroup = append(*funcGroup, &hookValue{args, scope})
}

func runHook(funcName string, args []string) error {
	switch funcName {
	case "enterScope":
		enterScope()
	case "exitScope":
		exitScope()
	case "returnToFileScope":
		returnToFileScope()
	case "enterFileScope":
		enterFileScope()
	case "exploreUsing":
		exploreUsing()
	case "writeManifest":
		writeManifest()
	case "assignManifest":
		assignManifest()
	default:
		return fmt.Errorf("No such function \"%s\"", funcName)
	}
	return nil
}

func runHooks(varName string, hookName string) error {
	varGroup, ok := variableHooks[varName]
	if !ok {
		// Just ignore when no hooks exist
		return nil
	}
	funcGroup, ok := (*varGroup)[hookName]
	if !ok {
		// Just ignore when no hooks exist
		return nil
	}
	for _, hook := range *funcGroup {
		if err := runHook(hook.args[0], hook.args[1:]); err != nil {
			return err
		}
	}
	return nil
}

func resolveString(str *Expression) (string, error) {
	comps := str.components
	out := ""
	if comps[0].id == stringId {
		targets := []string{}
		readingTarget := false
		for _, c := range comps[0].value {
			if readingTarget {
				if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' {
					targets[len(targets)-1] += string(c)
				} else {
					readingTarget = false
				}
			}
			// omit the % from target
			if c == '%' {
				readingTarget = true
				targets = append(targets, "")
			}
		}
		out = comps[0].value
		for _, varName := range targets {
			if rep, err := getVarOnTop(varName); err == nil {
				out = strings.ReplaceAll(out, "%"+varName, rep)
			} else {
				return "", err
			}
		}
	} else {
		for _, unit := range comps {
			if unit.id == wordId {
				out += " " + unit.value
			} else {
				varName := unit.value
				if rep, err := getVarOnTop(varName); err == nil {
					out += " " + rep
				} else {
					return "", err
				}
			}
		}
		out = out[1:]
	}
	return out, nil
}

func writeSong(original string, replacement string, manifestPath string, destFolder string) (string, error) {
	manifestFolder := filepath.Dir(manifestPath)
	srcExt := filepath.Ext(original)
	var srcPath string
	switch srcExt {
	case ".mmrs":
		srcPath = filepath.Join(manifestFolder, "mmrs", original)
	case ".zseq":
		fallthrough
	case ".aseq":
		fallthrough
	case ".seq":
		srcPath = filepath.Join(manifestFolder, "seq", original)
	default:
		return "", fmt.Errorf("%s is not a handled extension", srcExt)
	}

	if fs, err := os.Stat(srcPath); err != nil {
		return "", fmt.Errorf("file does not exist")
	} else if fs.IsDir() {
		return "", fmt.Errorf("is not a file")
	}

	destPath := filepath.Join(destFolder, replacement)
	if _, err := os.Stat(destPath); err == nil {
		// add author names to disambiguate
		repExt := filepath.Ext(replacement)
		repName := filepath.Base(replacement)
		repName = repName[0 : len(repName)-len(repExt)]
		authors, err := getVarOnTopHookless("seq")
		if err != nil {
			authors = "UNKNOWN"
		}
		authorList := []string{}
		for _, author := range strings.Split(authors, "&&") {
			authorList = append(authorList, strings.TrimSpace(author))
		}
		formattedAuthors := strings.Join(authorList, ",")
		destPath = filepath.Join(destFolder, fmt.Sprintf("%s (%s)%s", repName, formattedAuthors, repExt))
		if _, err := os.Stat(destPath); err == nil {
			// do not include the same song twice
			return "", fmt.Errorf("sequence by these authors already exists at the destination")
		}
	}
	for key := range bio {
		if val, err := getVarOnTopHookless(key); err == nil {
			bio[key] = val
		} else {
			bio[key] = ""
		}
	}
	err := mmrs.MakeCreditedArchive(srcPath, &destPath, &bio)
	if err != nil {
		return "", err
	}
	return filepath.Base(destPath), nil
}

func execute(program []*Statement, srcFilePath string, destFolder string) error {
	// unhandled error for now
	os.Mkdir(destFolder, os.ModePerm)
	for _, statement := range program {
		comps := statement.components
		switch statement.id {
		case assignmentId:
			varName := comps[0].components[0].value
			value, err := resolveString(comps[1])
			if err != nil {
				return err
			}
			putVarOnTop(varName, value)
		case modificationId:
			varName := comps[0].components[0].value
			hooks := comps[1:]
			for _, hook := range hooks {
				expComps := hook.components
				funcName := expComps[0].value
				argTokens := expComps[1:]
				argStrings := []string{}
				for _, arg := range argTokens {
					argStrings = append(argStrings, arg.value)
				}
				putHookOnTop(varName, funcName, argStrings)
			}
		case replacementId:
			rawReplacement, err := resolveString(comps[0])
			if err != nil {
				return err
			}
			replacementParts := strings.Split(rawReplacement, "::")
			if len(replacementParts) != 2 {
				return fmt.Errorf("Filename replacement at %d:%d is missing second argument", statement.line+1, statement.column+1)
			}
			original := strings.TrimSpace(replacementParts[0])
			repPrefix, err := getVarOnTopHookless("prefix")
			if err != nil {
				return err
			}
			repSuffix, err := getVarOnTopHookless("suffix")
			if err != nil {
				return err
			}
			seqAuthors, err := getVarOnTopHookless("seq")
			if err != nil {
				return err
			}
			replacement := fmt.Sprintf("%s%s (%s)%s", repPrefix, strings.TrimSpace(replacementParts[1]), seqAuthors, repSuffix)

			if songName, err := writeSong(original, replacement, srcFilePath, destFolder); err != nil {
				fmt.Printf("Could not write %s\n\tCause: %s\n", replacement, err)
			} else {
				currentManifest.value = fmt.Sprintf("%s\n\t%s", currentManifest.value, songName)
			}
		}
	}
	return nil
}
