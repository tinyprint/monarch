package monarch

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:embed template.lua.tmpl
var defaultMigrationTemplate string

var regexpMigMatchFileName = regexp.MustCompile("[0-9]{14}.*\\.lua")
var regexpMigMatchUnderscore = regexp.MustCompile("(_+)([a-zA-Z0-9])")

type files struct {
	directory string
}

func (f *files) templateFilePath() string {
	return path.Join(f.directory, "template.lua.tmpl")
}

func (f *files) validateDirectory() error {
	dir := f.directory
	fi, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	return nil
}

func (f *files) initDirectory() error {
	dir := f.directory
	fi, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if !fi.IsDir() {
		return fmt.Errorf("'%s' is not a directory", dir)
	}

	templateFile := f.templateFilePath()
	_, err = os.Stat(templateFile)
	if os.IsNotExist(err) {
		err = os.WriteFile(templateFile, []byte(defaultMigrationTemplate), 0644)
	} else if err != nil {
		return err
	}

	return nil
}

func (f *files) createNewMigrationFile(name string) error {
	datetime := time.Now().UTC().Format("20060102150405")
	migrationName := datetime + "_" + toCamelCase(name)
	fileName := path.Join(f.directory, migrationName+".lua")

	_, err := os.Stat(fileName)
	if !os.IsNotExist(err) {
		return fmt.Errorf("file %s already exists", fileName)
	}

	templateFilePath := f.templateFilePath()
	_, err = os.Stat(templateFilePath)
	if err != nil {
		return fmt.Errorf("template file %s not found", templateFilePath)
	}
	migrationTemplate, err := os.ReadFile(templateFilePath)
	if err != nil {
		return err
	}

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	tmpl, err := template.New("migration").Parse(string(migrationTemplate))
	if err != nil {
		return err
	}

	err = tmpl.Execute(
		file, struct {
			MigrationName string
		}{
			migrationName,
		},
	)

	err = file.Close()

	return err
}

func (f *files) getMigrationFiles() ([]string, error) {
	fi, err := os.Stat(f.directory)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, err
	}
	d, err := os.Open(f.directory)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	names, _ := d.Readdirnames(-1)
	sort.Strings(names)

	files := make([]string, 0, len(names))
	for _, n := range names {
		if regexpMigMatchFileName.Match([]byte(n)) {
			files = append(files, n)
		}
	}

	return files, nil
}

func (f *files) migrationPath(name string) string {
	return path.Join(f.directory, name)
}

func toCamelCase(name string) string {
	camel := regexpMigMatchUnderscore.ReplaceAllStringFunc(
		name, func(from string) string {
			return strings.ReplaceAll(strings.ToUpper(from), "_", "")
		},
	)

	return cases.Title(language.AmericanEnglish, cases.NoLower).String(camel)
}
