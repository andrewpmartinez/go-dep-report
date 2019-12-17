# go-dep-report
Lists the non-internal Go package dependancies imported into a project. It can output in CSV, YAML, and JSON. In
addition it attempts to pull in license information (which needs a bit of work).

# Installation

```
go get github.com/andrewpmartinez/go-dep-report
```

# Usage

All code for a project must be fetched. Either via `go get`, `go build`, or similar commands. This utility does not
fetch source. Package names will be search for using the standard Go lookup mechanisms; which will look for packages
in your `GOROOT` and `GOPATH`. This functionality is provided by Go's built in package import logic and will work in
in projects that use Go Modules and those that don't.

Some simple examples: 

```
go-dep-report ./my-go-project --depth 3 --format csv
```

```
go-dep-report github.com/someone/package
```


```
  go-dep-report <packageNameOrPath> [flags]
  go-dep-report [command]

Examples:
go-dep-report ./my-project

Available Commands:
  help        Help about any command
  version     Prints version information

Flags:
  -c, --config string       config file if desired (default is ~/.config/.go-dep-report)
  -d, --depth               the depth to resolve dependencies to (0 = no limit)
  -f, --format string       the output format to use (csv, json, yaml)
  -h, --help                help for go-dep-report
  -l, --log-format string   set the log format output (json, text)
  -o, --out-file string     set the file to route output to
  -v, --verbose             whether to increase log output or not

Use "go-dep-report [command] --help" for more information about a command.
```

# Configuration

All flags and options can be specified:

- in configuration files (least specific)
- in environment variables
- on the command line (most specific)

If the same flag/option is specified on multiple levels the most specific level will be used.

Environment variables are the uppercase flag names prefixed with `GDR_`

Example: `GDR_DEPTH`

Configuration files YAML with singular properties with the same flag name.

Example:

```yaml
depth: 2
format: csv
```


# License Report Information

Currently license information uses the first three lines of LICENSE files. If a repository doesn't have a LICENSE file
`???` will be used.