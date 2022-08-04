package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

var (
	bundle      = i18n.NewBundle(language.English)
	defaultLang = language.English
	localizer   = i18n.NewLocalizer(bundle, defaultLang.String())
)

//go:embed locales
var locales embed.FS

func init() {
	bundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.LoadMessageFileFS(locales, "locales/en.yml")
	bundle.LoadMessageFileFS(locales, "locales/zh.yml")
}

// Select default language
func Select(tag language.Tag) {
	defaultLang = tag
	localizer = i18n.NewLocalizer(bundle, defaultLang.String())
}

// Add language file or dir(auto get language by filename)
func Add(f string) {
	info, err := os.Stat(f)
	if err != nil {
		return
	}
	if info.IsDir() {
		filepath.Walk(f, func(path string, fi os.FileInfo, errBack error) error {
			if !fi.IsDir() {
				bundle.LoadMessageFile(path)
			}
			return nil
		})
	} else {
		bundle.LoadMessageFile(f)
	}
}

func T(id string) string {
	return localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID: id,
		},
	})
}

func E(id string) error {
	return fmt.Errorf(T(id))
}
