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
	sudoConfig = "/etc/pam.d/sudo_local"
	sudoPrefix = "/files/etc/pam.d/sudo_local"
)

func augeasConf() augeas.Flag {
	slog.Debug("no Augeas config parsing yet")
	return augeas.None
}

func handleError(msg string, err error) error {
	slog.Error(msg, "msg", err.Error())
	return err
}

func copyTemplate() error {
	src, err := os.Open(fmt.Sprintf("%s.template", sudoConfig))
	if err != nil {
		return handleError("unable to read sudoConfig template", err)
	}
	defer src.Close()

	dst, err := os.Create(sudoConfig)
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

func Run() error {
	// FIXME check for macOS >= Sonoma
	// https://stackoverflow.com/a/75955475/17597

	err := copyTemplate()
	if err != nil {
		return handleError("unable to copy template", err)
	}

	aug, err := augeas.New("/", "", augeasConf())
	if err != nil {
		return handleError("unable to initialize Augeas", err)
	}
	defer aug.Close()

	_, err = aug.DefineVariable("pam_tid", path.Join(sudoPrefix, "1"))
	if err != nil {
		return handleError("unable to define Augeas variable", err)
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

	return nil
}
