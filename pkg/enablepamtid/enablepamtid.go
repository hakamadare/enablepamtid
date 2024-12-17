package enablepamtid

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"

	"honnef.co/go/augeas"
)

// FIXME ugh
const (
	defaultSudoConfigPath = "/etc/pam.d/sudo_local"
	defaultAugeasPrefix   = "/files"
)

type config struct {
	sudoConfigPath   string
	augeasPathPrefix string
}

func augeasConf() augeas.Flag {
	slog.Debug("no Augeas config parsing yet")
	return augeas.None
}

func handleError(msg string, err error) error {
	slog.Error(msg, "msg", err.Error())
	return err
}

func copyTemplate(cfg *config) error {
	src, err := os.Open(fmt.Sprintf("%s.template", cfg.sudoConfigPath))
	if err != nil {
		return handleError("unable to read sudoConfig template", err)
	}
	defer src.Close()

	dst, err := os.Create(cfg.sudoConfigPath)
	if err != nil {
		return handleError("unable to create sudoConfig", err)
	}
	defer dst.Close()

	_, err = dst.ReadFrom(src)
	if err != nil {
		return handleError("unable to read from sudoConfig template", err)
	}

	_, err = io.Copy(dst, src)
	if err != nil {
		return handleError("unable to write sudoConfig", err)
	}

	return nil
}

func Config(cfgPath string, augPrefix string) *config {
	if cfgPath == "" {
		cfgPath = defaultSudoConfigPath
	}
	if augPrefix == "" {
		augPrefix = defaultAugeasPrefix
	}

	return &config{sudoConfigPath: cfgPath, augeasPathPrefix: augPrefix}
}

func Run(cfg *config) error {
	// FIXME check for macOS >= Sonoma
	// https://stackoverflow.com/a/75955475/17597

	err := copyTemplate(cfg)
	if err != nil {
		return handleError("unable to copy template", err)
	}

	aug, err := augeas.New("/", "", augeasConf())
	if err != nil {
		return handleError("unable to initialize Augeas", err)
	}
	defer aug.Close()

	_, created, err := aug.DefineNode(
		"pam_tid",
		path.Join(cfg.augeasPathPrefix, cfg.sudoConfigPath, "1"),
		"",
	)
	if err != nil {
		return handleError("unable to define Augeas variable", err)
	}
	if !created {
		return fmt.Errorf("no additional node created in sudo_local")
	}

	err = aug.Set(path.Join("$pam_tid", "type"), "auth")
	if err != nil {
		return handleError("unable to set sudo element type value", err)
	}

	err = aug.Set(path.Join("$pam_tid", "control"), "sufficient")
	if err != nil {
		return handleError("unable to set sudo element control value", err)
	}

	err = aug.Set(path.Join("$pam_tid", "module"), "pam_tid.so")
	if err != nil {
		return handleError("unable to set sudo element module value", err)
	}

	err = aug.Save()
	if err != nil {
		return handleError("unable to save pending changes", err)
	}

	errors, err := aug.Match(path.Join("/augeas", cfg.sudoConfigPath, "/error"))
	if err != nil {
		return handleError("unable to query Augeas errors", err)
	}
	for _, v := range errors {
		slog.Error("augeas error", "err", v)
	}

	return nil
}
