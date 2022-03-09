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
	kustomizePluginConfigRoot string

	fsOnDisk        filesys.FileSystem
	kustomizer      *krusty.Kustomizer
	patternMatchers = map[directoryBase]*fileutils.PatternMatcher{}

	gitRootPath    string
	gitRootPathLen int
)

type KustomizeBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec Spec `json:"spec,omitempty"`
}

type Spec struct {
	Directories []Directory `json:"directories,omitempty"`
}

type Directory struct {
	Base string   `json:"base,omitempty"`
	Glob []string `json:"glob,omitempty"`
}

func main() {
	filePath := os.Args[1]

	env, exists := os.LookupEnv(kustomizePluginConfigRootEnv)
	if !exists {
		log.Panicf("%s is empty", kustomizePluginConfigRootEnv)
	}
	kustomizePluginConfigRoot = env

	fsOnDisk = filesys.MakeFsOnDisk()
	kustomizer = makeKustomizer()

	if err := makePatternMatchers(filePath); err != nil {
		log.Panic(filePath, ": ", err)
	}

	gitRootPathValue, err := getGitRootPath(kustomizePluginConfigRoot)
	if err != nil {
		log.Panic(filePath, ": ", err)
	}
	gitRootPath = gitRootPathValue
	gitRootPathLen = len(gitRootPath)

	if err = fsOnDisk.Walk(gitRootPath, walk); err != nil {
		log.Panic(filePath, ": ", err)
	}
}

func makeKustomizer() *krusty.Kustomizer {
	krustyOptions := krusty.MakeDefaultOptions()
	krustyOptions.PluginConfig = types.EnabledPluginConfig(types.BploUseStaticallyLinked)
	return krusty.MakeKustomizer(krustyOptions)
}

func makePatternMatchers(filePath string) error {
	gitPatternMatcher, err := makePatternMatcher(git, filePath)
	if err != nil {
		return err
	}
	patternMatchers[git] = gitPatternMatcher

	pwdPatternMatcher, err := makePatternMatcher(pwd, filePath)
	if err != nil {
		return err
	}
	patternMatchers[pwd] = pwdPatternMatcher

	return nil
}

func makePatternMatcher(base directoryBase, filePath string) (*fileutils.PatternMatcher, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var kustomizeBuild KustomizeBuild
	if err := yaml.Unmarshal(data, &kustomizeBuild); err != nil {
		return nil, err
	}

	var sb strings.Builder
	for _, d := range kustomizeBuild.Spec.Directories {
		if d.Base != base.string() {
			continue
		}

		for _, g := range d.Glob {
			sb.WriteString(g)
			sb.WriteString("\n")
		}
	}

	patterns, err := dockerignore.ReadAll(strings.NewReader(sb.String()))
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

func walk(path string, info fs.FileInfo, err error) error {
	if err != nil || !info.IsDir() {
		return err
	}

	for directoryBase, patternMatcher := range patternMatchers {
		matchPath, err := directoryBase.path(path)
		if err != nil {
			return err
		}
		matches, err := patternMatcher.Matches(matchPath)
		if err != nil {
			return err
		}
		if matches {
			if err := runKustomizeBuild(path); err != nil {
				return err
			}
		}
	}

	return nil
}

func runKustomizeBuild(path string) error {
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

	_, err = os.Stdout.Write([]byte(yamlSeparator))

	return err
}
