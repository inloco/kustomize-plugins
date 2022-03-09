package main

import (
	"fmt"
	"path/filepath"
)

type directoryBase int

const (
	git directoryBase = iota
	pwd
)

func (b directoryBase) string() string {
	switch b {
	case git:
		return "git"
	case pwd:
		return "pwd"
	default:
		panic(fmt.Sprintf("unknown directory base type: %d", b))
	}
}

func (b directoryBase) path(path string) (string, error) {
	switch b {
	case git:
		if path == gitRootPath {
			return ".", nil
		}
		return path[gitRootPathLen+1:], nil
	case pwd:
		return filepath.Rel(kustomizePluginConfigRoot, path)
	default:
		return "", fmt.Errorf("unknown directory base type: %d", b)
	}
}
