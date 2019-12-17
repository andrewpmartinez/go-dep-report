package godepreport

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"os"
)

type Context interface {
	Packages() []string
	Log() *logrus.Logger
	Formatter() Formatter
	Writer() io.Writer
	Close()
	Depth() int
}

type Formatter interface {
	AddEntry(ctx Context, entry Entry)
	Write(ctx Context)
}

type formattingEntry struct {
	Parent  string `json:"parent" yaml:"parent"`
	Package string `json:"package" yaml:"package"`
	License string `json:"license" yaml:"license"`
}

type FormatterJSON struct {
	entries []formattingEntry
}

func (formatter *FormatterJSON) Write(ctx Context) {
	entryJson, err := json.MarshalIndent(formatter.entries, "", "  ")
	if err != nil {
		ctx.Log().WithError(err).Error("could not encode formatting entry to JSON")
	}
	_, err = ctx.Writer().Write(entryJson)
	if err != nil {
		ctx.Log().WithError(err).Error("could not write JSON")
	}
}

func (formatter *FormatterJSON) AddEntry(ctx Context, entry Entry) {
	formatter.entries = append(formatter.entries, formattingEntry{
		Parent:  entry.ParentPkgName(),
		Package: entry.PkgName(),
		License: entry.LicenseName(),
	})
}

type FormatterCSV struct {
	lines []string
}

func (formatter *FormatterCSV) Write(ctx Context) {
	_, err := ctx.Writer().Write([]byte(fmt.Sprintf("%s,%s,%s\n", "Parent", "Package", "License")))

	if err != nil {
		ctx.Log().Error("could not write CSV Header: %v", err)
	}

	for _, line := range formatter.lines {
		_, err = ctx.Writer().Write([]byte(line))

		if err != nil {
			ctx.Log().Error("could not write CSV line: %v", err)
		}
	}
}

func (formatter *FormatterCSV) AddEntry(ctx Context, entry Entry) {
	line := fmt.Sprintf("%s,%s,%s\n", entry.ParentPkgName(), entry.PkgName(), entry.LicenseName())

	formatter.lines = append(formatter.lines, line)
}

type FormatterYAML struct {
	entries []formattingEntry
}

func (formatter *FormatterYAML) Write(ctx Context) {
	entryYaml, err := yaml.Marshal(formatter.entries)
	if err != nil {
		ctx.Log().WithError(err).Error("could not encode formatting entry to YAML")
	}
	_, err = ctx.Writer().Write(entryYaml)
	if err != nil {
		ctx.Log().WithError(err).Error("could not write YAML")
	}
}

func (formatter *FormatterYAML) AddEntry(ctx Context, entry Entry) {
	formatter.entries = append(formatter.entries, formattingEntry{
		Parent:  entry.ParentPkgName(),
		Package: entry.PkgName(),
		License: entry.LicenseName(),
	})
}

type Entry interface {
	ParentPkgName() string
	PkgName() string
	LicenseName() string
}

func Run(ctx Context) {
	t := &Tree{
		ResolveInternal: false,
		ResolveTest:     false,
		MaxDepth:        ctx.Depth(),
	}

	for _, pkg := range ctx.Packages() {

		err := t.Resolve(ctx, pkg)
		if err != nil {
			cwd, _ := os.Getwd()
			ctx.Log().WithField("currentDirectory", cwd).Fatal(err)
		}

	}

	format(ctx, *t.Root)
	ctx.Formatter().Write(ctx)
}

func format(ctx Context, root Pkg) {
	formatDeps(ctx, root.Deps)

	for _, dep := range root.Deps {
		format(ctx, dep)
	}
}

func formatDeps(ctx Context, deps []Pkg) {
	for _, dep := range deps {
		if !dep.Internal {
			ctx.Formatter().AddEntry(ctx, &dep)
		}

	}
}
