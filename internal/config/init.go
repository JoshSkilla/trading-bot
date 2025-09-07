package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func init() {
	root := findModuleRoot()
	if root != "" {
		_ = godotenv.Load(filepath.Join(root, ".env"))
	} else {
		panic(fmt.Errorf("could not find module root"))
	}
}

	// Climbs up the directory tree up to 10 levels to find root with go.mod
func findModuleRoot() string {
	dir, _ := os.Getwd()
	for i := 0; i < 10 && dir != "/" && dir != ""; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return ""
}