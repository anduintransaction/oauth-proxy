//go:generate goru packer template/app generated/app
package creator

import (
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gottb.io/goru"
	"gottb.io/goru/errors"
	"gottb.io/goru/packer"
	"gottb.io/goru/tool/creator/generated/app"
	"gottb.io/goru/utils"
)

func AppCreator(args []string, options map[string]string) {
	if len(args) == 0 {
		goru.ErrPrintln(goru.ColorWhite, "USAGE: goru create app <appname>")
		os.Exit(1)
	}
	appName := args[0]
	matched, _ := regexp.MatchString("^[0-9a-zA-Z\\-_\\.]+$", appName)
	if !matched {
		goru.ErrPrintln(goru.ColorRed, "Invalid app name")
		os.Exit(1)
	}
	_, err := os.Stat(appName)
	if err == nil {
		goru.ErrPrintln(goru.ColorRed, "App folder existed")
		os.Exit(1)
	}
	err = os.MkdirAll(appName, 0755)
	if err != nil {
		goru.ErrPrintln(goru.ColorWhite, err)
		os.Exit(1)
	}
	appPackage, err := utils.PkgPath(appName)
	if err != nil {
		goru.ErrPrintln(goru.ColorWhite, err)
		os.Exit(1)
	}
	filelist := app.VVVPack.List()
	randomString := createRandomString()
	for _, filename := range filelist {
		err = createAppFile(appPackage, appName, filename, randomString)
		if err != nil {
			goru.ErrPrintln(goru.ColorRed, err)
			os.RemoveAll(appName)
			os.Exit(1)
		}
	}
	goru.Printf(goru.ColorGreen, "App %q created successfully\n", appName)
}

func createAppFile(appPackage, appName, filename, randomString string) error {
	path := strings.TrimSuffix(filepath.Join(appName, filepath.Clean(filename)), ".tmpl")
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return errors.Wrap(err)
	}
	if filename == "public/empty" {
		return nil
	}
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err)
	}
	defer f.Close()
	f.Chmod(0644)
	tmpl, err := packer.LoadTemplate(app.VVVPack, filename)
	if err != nil {
		return err
	}
	tmplData := &struct {
		BinName   string
		PublicPkg string
		Secret    string
	}{
		BinName:   "___" + appName,
		PublicPkg: appPackage + "/generated/assets/public",
		Secret:    randomString,
	}
	return errors.Wrap(tmpl.Execute(f, tmplData))
}

func createRandomString() string {
	rand.Seed(time.Now().UnixNano())
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 32)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
