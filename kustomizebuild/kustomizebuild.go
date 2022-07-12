package main

import (
	"fmt"
	"io"
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
	panicSeparator = ": "
	yamlSeparator  = "---\n"

	kustomizePluginConfigRootEnv = "KUSTOMIZE_PLUGIN_CONFIG_ROOT"
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

func (b directoryBase) parsePath(gitRootPath string, kustomizationPath string, path string) (string, error) {
	switch b {
	case git:
		if path == gitRootPath {
			return ".", nil
		}
		return path[len(gitRootPath)+1:], nil
	case pwd:
		return filepath.Rel(kustomizationPath, path)
	default:
		return "", fmt.Errorf("unknown directory base type: %d", b)
	}
}

type KustomizeBuild struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              Spec `json:"spec,omitempty"`
}

type Spec struct {
	Directories []Directory `json:"directories,omitempty"`
}

type Directory struct {
	Base  string   `json:"base,omitempty"`
	Globs []string `json:"globs,omitempty"`
}

func main() {
	filePath := os.Args[1]

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Panic(filePath, panicSeparator, err)
	}

	if err := GenerateManifests(data, os.Stdout); err != nil {
		log.Panic(filePath, panicSeparator, err)
	}
}

func GenerateManifests(data []byte, out io.Writer) error {
	var kustomizeBuild KustomizeBuild
	if err := yaml.Unmarshal(data, &kustomizeBuild); err != nil {
		return err
	}

	manifests, err := makeManifests(&kustomizeBuild)
	if err != nil {
		return err
	}

	for _, manifest := range manifests {
		if _, err := out.Write([]byte(yamlSeparator)); err != nil {
			return err
		}

		if _, err := out.Write(manifest); err != nil {
			return err
		}
	}

	return nil
}

func makeManifests(kustomizeBuild *KustomizeBuild) ([][]byte, error) {
	patternMatchers, err := makePatternMatchers(kustomizeBuild)
	if err != nil {
		return nil, err
	}

	manifests, err := runKustomizations(patternMatchers)
	if err != nil {
		return nil, err
	}

	return manifests, nil
}

func makePatternMatchers(kustomizeBuild *KustomizeBuild) (map[directoryBase]*fileutils.PatternMatcher, error) {
	patternMatchers := make(map[directoryBase]*fileutils.PatternMatcher)

	gitPatternMatcher, err := makePatternMatcher(git, kustomizeBuild)
	if err != nil {
		return nil, err
	}
	patternMatchers[git] = gitPatternMatcher

	pwdPatternMatcher, err := makePatternMatcher(pwd, kustomizeBuild)
	if err != nil {
		return nil, err
	}
	patternMatchers[pwd] = pwdPatternMatcher

	return patternMatchers, nil
}

func makePatternMatcher(dirBase directoryBase, kustomizeBuild *KustomizeBuild) (*fileutils.PatternMatcher, error) {
	var sb strings.Builder

	for _, dir := range kustomizeBuild.Spec.Directories {
		if dir.Base == dirBase.string() {
			for _, glob := range dir.Globs {
				sb.WriteString(glob)
				sb.WriteString("\n")
			}
		}
	}

	patterns, err := dockerignore.ReadAll(strings.NewReader(sb.String()))
	if err != nil {
		return nil, err
	}

	return fileutils.NewPatternMatcher(patterns)
}

func runKustomizations(patternMatchers map[directoryBase]*fileutils.PatternMatcher) ([][]byte, error) {
	fileSystem := filesys.MakeFsOnDisk()

	kustomizationPath, exists := os.LookupEnv(kustomizePluginConfigRootEnv)
	if !exists {
		return nil, fmt.Errorf("%s is empty", kustomizePluginConfigRootEnv)
	}

	gitRootPath, err := getGitRootPath(fileSystem, kustomizationPath)
	if err != nil {
		return nil, err
	}

	kustomizer := makeKustomizer()

	var manifests [][]byte

	if err := fileSystem.Walk(gitRootPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return err
		}

		for dirBase, patternMatcher := range patternMatchers {
			matchPath, err := dirBase.parsePath(gitRootPath, kustomizationPath, path)
			if err != nil {
				return err
			}

			matches, err := patternMatcher.Matches(matchPath)
			if err != nil {
				return err
			}

			if matches {
				resMap, err := kustomizer.Run(fileSystem, path)
				if err != nil {
					return err
				}
				b, err := resMap.AsYaml()
				if err != nil {
					return err
				}
				manifests = append(manifests, b)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return manifests, nil
}

func makeKustomizer() *krusty.Kustomizer {
	krustyOptions := krusty.MakeDefaultOptions()
	krustyOptions.PluginConfig = types.EnabledPluginConfig(types.BploUseStaticallyLinked)

	return krusty.MakeKustomizer(krustyOptions)
}

func getGitRootPath(fileSystem filesys.FileSystem, kustomizationPath string) (string, error) {
	path := kustomizationPath
	for path != "/" {
		path = filepath.Dir(path)

		if fileSystem.IsDir(filepath.Join(path, ".git")) {
			return path, nil
		}
	}

	return "", fmt.Errorf("unable to find git root in '%s' parents", kustomizationPath)
}
