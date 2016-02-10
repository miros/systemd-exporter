package systemd

import (
  "systemd-exporter/systemd/validation"
  "github.com/imdario/mergo"
  "github.com/spf13/afero"
)

import "github.com/davecgh/go-spew/spew"
var _ = spew.Dump

func (sys *Systemd) Install(appName string, services []Service) {
  setServiceDefaults(services, sys.Config)
  validateParams(appName, sys.Config, services)
  sys.doInstall(appName, services)
}

func (sys *Systemd) doInstall(appName string, services []Service) {
  sys.installServices(appName, services)
  sys.writeAppUnit(appName, services)
  sys.MustEnableService(appName)
}

func validateParams(appName string, config Config, services []Service) {
  validateAppName(appName)
  validateConfig(config)
  validateServices(services)
}

func validateAppName(appName string) {
  if err := validation.NoSpecialSymbols(appName); err != nil {
    panic(err)
  }
}

func validateConfig(config Config) {
  validation.MustBeValid(&config)
}

func setServiceDefaults(services []Service, config Config) {
  for i, _ := range services {
    defaults := ServiceOptions{User: config.User, Group: config.Group, WorkingDirectory: config.DefaultWorkingDirectory}
    mergo.Merge(&services[i].Options, defaults)
  }
}

func validateServices(services []Service) {
  for _, service := range(services) {
    validation.MustBeValid(&service)
  }
}

func (sys *Systemd) installServices(appName string, services []Service) {
  error := sys.fs.MkdirAll(sys.Config.HelperDir, 0755)
  if error != nil {
    panic(error)
  }

  for _, service := range(services) {
    sys.writeServiceUnit(appName, service)
  }
}

func (sys *Systemd) writeAppUnit(appName string, services []Service) {
  path := sys.Config.unitPath(appName)
  data := RenderAppTemplate(appName, sys.Config, services)
  writeFile(sys.fs, path, data)
}

func (sys *Systemd) writeServiceUnit(appName string, service Service) {
  fullServiceName := service.fullName(appName)

  service.helperPath = sys.Config.helperPath(fullServiceName)
  helperData := RenderHelperTemplate(service.Cmd)
  writeFile(sys.fs, service.helperPath, helperData)

  unitPath := sys.Config.unitPath(fullServiceName)
  writeFile(sys.fs, unitPath, RenderServiceTemplate(appName, service))
}

func writeFile(fs afero.Fs, path string, data string) {
  error := afero.WriteFile(fs, path, []byte(data), 0644)
  if error != nil {
    panic(error)
  }
}
