package systemd

import (
  "systemd-exporter/systemd/validation"
)

type Respawn struct {
  Count int
  Interval int
}

type ServiceOptions struct {
  WorkingDirectory string
  Env map[string]string
  User string
  Group string
  KillTimeout int
  Respawn Respawn
}

func (options *ServiceOptions) Validate() error {
  if err := validation.Path(options.WorkingDirectory); err != nil {
    return err
  }

  if err := validation.NoSpecialSymbols(options.User); err != nil {
    return err
  }

  if err := validation.NoSpecialSymbols(options.Group); err != nil {
    return err
  }

  return nil
}

type Service struct {
  Name string
  Cmd string
  Options ServiceOptions
  helperPath string
}

func (service *Service) Validate() error {
  if err := validation.NoSpecialSymbols(service.Name); err != nil {
    return err
  }

  if err := service.Options.Validate(); err != nil {
    return err
  }

  return nil
}

func (service *Service) fullName(appName string) string {
  return appName + "_" + service.Name
}
