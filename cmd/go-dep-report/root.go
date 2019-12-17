package main

import (
	"fmt"
	godepreport "github.com/andrewpmartinez/go-dep-reporter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"os"
	"strings"
)

type context struct {
	viper     *viper.Viper
	log       *logrus.Logger
	formatter godepreport.Formatter
	writer    io.Writer
	close     func()
	options   *options
}

func (ctx *context) Depth() int {
	return ctx.viper.GetInt("depth")
}

func (ctx *context) Packages() []string {
	return ctx.viper.GetStringSlice("packages")
}
func (ctx *context) Log() *logrus.Logger {
	return ctx.log
}
func (ctx *context) Formatter() godepreport.Formatter {
	return ctx.formatter
}
func (ctx *context) Writer() io.Writer {
	return ctx.writer
}

func (ctx *context) Close() {
	if ctx.close != nil {
		ctx.close()
	}
}

type options struct {
	verbose   bool
	config    string
	format    string
	logFormat string
	packages  []string
	outFile   string
	depth     int
}

func main() {
	ctx := &context{
		options: &options{
			verbose: false,
			config:  "",
		},
		log:   logrus.New(),
		viper: viper.New(),
	}

	ctx.viper.SetConfigName("go-dep-report.yml")
	ctx.viper.SetConfigType("yaml")
	ctx.viper.AddConfigPath(".")

	if homedir, err := os.UserHomeDir(); err == nil {
		ctx.viper.AddConfigPath(homedir + "/.config")
	}

	ctx.viper.SetEnvPrefix("GDR")
	ctx.viper.AutomaticEnv()

	rootCmd := &cobra.Command{
		Use:     "go-dep-report <packageNameOrPath>",
		Example: "go-dep-report ./my-project/...",
		Short:   "A tool to report on golang project dependencies",
		Long:    "go-dep-report generates human and machine readable reports of golang project dependencies. It is meant to provide a relatively easy path of reporting open source library usage.",
		Args: func(cmd *cobra.Command, args []string) error {

			if len(args) != 1 {
				return fmt.Errorf("expected 1 or more package names, got %d", len(args))
			}
			ctx.viper.Set("packages", args)

			if ctx.viper.GetBool("verbose") {
				ctx.log.SetLevel(logrus.DebugLevel)
				ctx.log.Debug("verbose on")
			}

			logFormat := strings.ToLower(ctx.viper.GetString("log-format"))
			switch {
			case logFormat == "text":
				ctx.log.SetFormatter(&logrus.TextFormatter{
					ForceColors:               true,
					DisableColors:             false,
					EnvironmentOverrideColors: true,
					DisableTimestamp:          false,
					FullTimestamp:             true,
				})
			case logFormat == "json":
				//do nothing, default for logrus
			default:
				return fmt.Errorf("invalid log format specified [%s] valid values are [text, json]", logFormat)
			}

			outFile := strings.ToLower(ctx.viper.GetString("out-file"))
			switch {
			case outFile != "":
				file, err := os.Create(outFile)
				if err != nil {
					return err
				}

				ctx.writer = file
				ctx.close = func() {
					_ = file.Close()
				}
			default:
				ctx.writer = os.Stdout
			}

			format := strings.ToLower(ctx.viper.GetString("format"))
			switch {
			case format == "csv":
				ctx.formatter = &godepreport.FormatterCSV{}
			case format == "json":
				ctx.formatter = &godepreport.FormatterJSON{}
			case format == "yaml":
				ctx.formatter = &godepreport.FormatterYAML{}
			default:
				return fmt.Errorf("invalid format specified [%s] valid values are [csv, json, yaml]", format)
			}

			if config := ctx.viper.GetString("config"); config != "" {
				ctx.viper.SetConfigFile(config)
			}

			if err := ctx.viper.ReadInConfig(); err != nil {
				return err
			}

			ctx.log.Debug("config file used: " + ctx.viper.ConfigFileUsed())

			return nil
		},
		Run:
		func(cmd *cobra.Command, args []string) {
			godepreport.Run(ctx)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&ctx.options.config, "config", "c", "", "config file if desired (default is ~/.config/.go-dep-report)")
	ctx.viper.SetDefault("config", "")
	if err := ctx.viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config")); err != nil {
		panic("could not find flag [config]")
	}

	ctx.viper.SetDefault("verbose", false)
	rootCmd.PersistentFlags().BoolVarP(&ctx.options.verbose, "verbose", "v", false, "whether to increase log output or not")
	if err := ctx.viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		panic("could not find flag [verbose]")
	}

	ctx.viper.SetDefault("format", "csv")
	rootCmd.PersistentFlags().StringVarP(&ctx.options.format, "format", "f", "", "the output format to use (csv, json, yaml)")
	if err := ctx.viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format")); err != nil {
		panic("could not find flag [format]")
	}

	ctx.viper.SetDefault("log-format", "text")
	rootCmd.PersistentFlags().StringVarP(&ctx.options.logFormat, "log-format", "l", "", "set the log format output (json, text)")
	if err := ctx.viper.BindPFlag("log-format", rootCmd.PersistentFlags().Lookup("log-format")); err != nil {
		panic("could not find flag [log-format]")
	}

	ctx.viper.SetDefault("out-file", "")
	rootCmd.PersistentFlags().StringVarP(&ctx.options.outFile, "out-file", "o", "", "set the file to route output to")
	if err := ctx.viper.BindPFlag("out-file", rootCmd.PersistentFlags().Lookup("out-file")); err != nil {
		panic("could not find flag [out-file]")
	}

	ctx.viper.SetDefault("depth", 0)
	rootCmd.PersistentFlags().IntVarP(&ctx.options.depth, "depth", "d", 0, "the depth to resolve dependencies to (0 = no limit)")
	if err := ctx.viper.BindPFlag("depth", rootCmd.PersistentFlags().Lookup("depth")); err != nil {
		panic("could not find flag [depth]")
	}

	rootCmd.AddCommand(newVersionCmd())

	_ = rootCmd.Execute()
}

func newVersionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Prints version information",
		Run: func(cmd *cobra.Command, args []string) {
			println("go-dep-report " + godepreport.Version())
		},
	}

	return versionCmd
}
