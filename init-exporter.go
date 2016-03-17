package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/miros/init-exporter/exporter"
	"github.com/miros/init-exporter/procfile"
	"github.com/miros/init-exporter/systemd"
	"github.com/miros/init-exporter/upstart"
	"github.com/miros/init-exporter/utils"
	"os"
)

import "github.com/davecgh/go-spew/spew"

var _ = spew.Dump

const version = "0.0.2"
const defaultConfigPath = "/etc/init-exporter.yaml"

const SYSTEMD = "systemd"
const UPSTART = "upstart"

func main() {
	defer prettyPrintPanics()

	app := cli.NewApp()
	describeApp(app, version)
	app.Action = runAction
	app.Run(os.Args)
}

func prettyPrintPanics() {
	if os.Getenv("DEBUG") == "true" {
		return
	}

	if r := recover(); r != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", r)
		os.Exit(1)
	}
}

func describeApp(app *cli.App, version string) {
	app.Name = "init-exporter"
	app.Usage = "exports services described by Procfile to systemd"
	app.Version = version

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "n, application_name",
			Usage: "Application name (This name only affects the names of generated files)",
		},
		cli.BoolFlag{
			Name:  "c, uninstall",
			Usage: "Remove scripts and helpers for a particular application",
		},
		cli.StringFlag{
			Name:  "config",
			Value: defaultConfigPath,
			Usage: "path to configuration file",
		},
		cli.StringFlag{
			Name:  "p, procfile",
			Usage: "path to procfile",
		},
	}
}

func runAction(cliContext *cli.Context) {
	appName := cliContext.String("application_name")

	if appName == "" {
		panic("No application name specified")
		return
	}

	globalConfig := ReadGlobalConfig(cliContext.String("config"))
	appName = globalConfig.Prefix + appName

	provider := newProvider(detectProvider(cliContext))
	exporter := newExporter(globalConfig, provider)

	if cliContext.Bool("uninstall") {
		uninstall(exporter, appName)
	} else {
		install(exporter, appName, cliContext.String("procfile"))
	}
}

func newExporter(config GlobalConfig, provider exporter.Provider) *exporter.Exporter {
	exporterConfig := exporter.Config{
		HelperDir: config.HelperDir,
		TargetDir: utils.TakeFirstDefined(config.TargetDir, provider.DefaultTargetDir()),
		User:      config.RunUser,
		Group:     config.RunGroup,
		DefaultWorkingDirectory: config.WorkingDirectory,
	}

	spew.Dump(config)

	return exporter.New(exporterConfig, provider)
}

func uninstall(exporter *exporter.Exporter, appName string) {
	exporter.Uninstall(appName)
	fmt.Println("systemd service uninstalled")
}

func install(exporter *exporter.Exporter, appName string, pathToProcfile string) {
	if pathToProcfile == "" {
		panic("No procfile given")
	}

	if services, err := procfile.ReadProcfile(pathToProcfile); err == nil {
		exporter.Install(appName, services)
		fmt.Println("systemd service installed to", exporter.Config.TargetDir)
	} else {
		panic(err)
	}
}

func newProvider(providerName string) exporter.Provider {
	switch providerName {
	case SYSTEMD:
		return systemd.New()
	case UPSTART:
		return upstart.New()
	default:
		panic("unknown init provider")
	}
}

func detectProvider(cliContext *cli.Context) string {
	return UPSTART
}
