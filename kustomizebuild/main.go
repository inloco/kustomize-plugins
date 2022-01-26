package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/dockerignore"
	"github.com/moby/moby/pkg/fileutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/yaml"
)

const (
	yamlSeparator                = "---\n"
	kustomizePluginConfigRootEnv = "KUSTOMIZE_PLUGIN_CONFIG_ROOT"
)

var (
	fsOnDisk filesys.FileSystem

	patternMatcher *fileutils.PatternMatcher

	gitRootPath    string
	gitRootPathLen int

	kustomizer *krusty.Kustomizer
)

type KustomizeBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KustomizeBuildSpec `json:"spec,omitempty"`
}

type KustomizeBuildSpec struct {
	Directories []string `json:"directories,omitempty"`
}

func main() {
	filePath := os.Args[1]

	kustomizePluginConfigRoot, ok := os.LookupEnv(kustomizePluginConfigRootEnv)
	if !ok {
		log.Panic(filePath, ": ", fmt.Errorf("%s is not set", kustomizePluginConfigRootEnv))
	}

	matcher, err := makePatternMatcher(filePath)
	if err != nil {
		log.Panic(filePath, ": ", err)
	}
	patternMatcher = matcher

	fsOnDisk = filesys.MakeFsOnDisk()

	path, err := getGitRootPath(kustomizePluginConfigRoot)
	if err != nil {
		log.Panic(filePath, ": ", err)
	}
	gitRootPath = path

	gitRootPathLen = len(gitRootPath)

	kustomizer = makeKustomizer()

	if err := fsOnDisk.Walk(gitRootPath, walk); err != nil {
		log.Panic(filePath, ": ", err)
	}
}

func makePatternMatcher(filePath string) (*fileutils.PatternMatcher, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var kustomizeBuild KustomizeBuild
	if err := yaml.Unmarshal(data, &kustomizeBuild); err != nil {
		return nil, err
	}

	reader := strings.NewReader(strings.Join(kustomizeBuild.Spec.Directories, "\n"))
	patterns, err := dockerignore.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return fileutils.NewPatternMatcher(patterns)
}

func getGitRootPath(filePath string) (string, error) {
	path := filepath.Dir(filePath)
	for {
		if fsOnDisk.IsDir(filepath.Join(path, ".git")) {
			return path, nil
		}

		if path == "/" {
			return "", fmt.Errorf("unable to find git root in '%s' parents", filePath)
		}

		path = filepath.Dir(path)
	}
}

func makeKustomizer() *krusty.Kustomizer {
	krustyOptions := krusty.MakeDefaultOptions()
	krustyOptions.PluginConfig = types.EnabledPluginConfig(types.BploUseStaticallyLinked)

	return krusty.MakeKustomizer(krustyOptions)
}

func walk(path string, info fs.FileInfo, err error) error {
	if err != nil || !info.IsDir() {
		return err
	}

	matchPath := "."
	if path != gitRootPath {
		matchPath = path[gitRootPathLen+1:]
	}

	matches, err := patternMatcher.Matches(matchPath)
	if err != nil {
		return err
	}
	if matches {
		m, err := kustomizer.Run(fsOnDisk, path)
		if err != nil {
			return err
		}
		b, err := m.AsYaml()
		if err != nil {
			return err
		}
		if _, err := os.Stdout.Write(b); err != nil {
			return err
		}

		if _, err := os.Stdout.Write([]byte(yamlSeparator)); err != nil {
			return err
		}
	}

	return nil
}
