package main_test

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"

	"github.com/inloco/iac-kustomize-plugins/kustomizebuild"
)

const (
	kustomizeBuildDir            = "k8s"
	kustomizationFileName        = "kustomization.yaml"
	kustomizePluginConfigRootEnv = "KUSTOMIZE_PLUGIN_CONFIG_ROOT"
)

var (
	separatorYaml = regexp.MustCompile("\n---\n")

	configMapGVK = v1.SchemeGroupVersion.WithKind(reflect.TypeOf(v1.ConfigMap{}).Name())
)

var _ = ginkgo.Describe("KustomizeBuild", func() {
	workingDir, err := os.MkdirTemp("", "*")
	g.Expect(err).To(g.BeNil())

	g.Expect(os.Mkdir(filepath.Join(workingDir, ".git"), 0700)).To(g.BeNil())
	g.Expect(os.Setenv(kustomizePluginConfigRootEnv, filepath.Join(workingDir, kustomizeBuildDir))).To(g.BeNil())

	kustomizationDirs := []string{
		"a/api",
		"a/app",
		"b/api",
	}
	g.Expect(generateKustomizations(workingDir, kustomizationDirs)).To(g.BeNil())

	ginkgo.DescribeTable("", KustomizeBuild,
		ginkgo.Entry("with git base",
			makeKustomizeBuild([]main.Directory{{
				Base: "git",
				Globs: []string{
					"a/**",
					"!a/app",
					"b/api",
				},
			}}),
			[]string{
				"a-api",
				"b-api",
			},
		),
		ginkgo.Entry("with pwd base",
			makeKustomizeBuild([]main.Directory{{
				Base: "pwd",
				Globs: []string{
					"../a/**",
					"!../a/app",
					"../b/api",
				},
			}}),
			[]string{
				"a-api",
				"b-api",
			},
		),
		ginkgo.Entry("with multiple base directories",
			makeKustomizeBuild([]main.Directory{
				{
					Base: "git",
					Globs: []string{
						"a/api",
					},
				},
				{
					Base: "pwd",
					Globs: []string{
						"../b/api",
					},
				},
			}),
			[]string{
				"a-api",
				"b-api",
			},
		),
	)
})

func generateKustomizations(workingDir string, kustomizationDirs []string) error {
	for _, kustomizationDir := range kustomizationDirs {
		kustomization := types.Kustomization{
			TypeMeta: types.TypeMeta{
				APIVersion: types.KustomizationVersion,
				Kind:       types.KustomizationKind,
			},
			ConfigMapGenerator: []types.ConfigMapArgs{{
				GeneratorArgs: types.GeneratorArgs{
					Name: strings.ReplaceAll(kustomizationDir, "/", "-"),
				}},
			},
		}

		filePath := filepath.Join(workingDir, kustomizationDir, kustomizationFileName)

		if err := os.MkdirAll(filepath.Dir(filePath), 0700); err != nil {
			return err
		}

		data, err := yaml.Marshal(kustomization)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

func makeKustomizeBuild(directories []main.Directory) main.KustomizeBuild {
	return main.KustomizeBuild{
		TypeMeta: metav1.TypeMeta{
			APIVersion: schema.GroupVersion{
				Group:   "incognia.com",
				Version: "v1alpha1",
			}.String(),
			Kind: "KustomizeBuild",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "_",
		},
		Spec: main.Spec{
			Directories: directories,
		},
	}
}

func KustomizeBuild(kustomizeBuild main.KustomizeBuild, expectedConfigMapNames []string) {
	kustomizeBuildYaml, err := yaml.Marshal(kustomizeBuild)
	g.Expect(err).To(g.BeNil())

	ginkgo.By("contains only expected GKVs", func() {
		var out bytes.Buffer
		g.Expect(main.GenerateManifests(kustomizeBuildYaml, &out)).To(g.Succeed())

		var actualGVKs []schema.GroupVersionKind
		for _, manifest := range separatorYaml.Split(out.String(), -1) {
			var typeMeta metav1.TypeMeta
			g.Expect(yaml.Unmarshal([]byte(manifest), &typeMeta)).To(g.Succeed())
			actualGVKs = append(actualGVKs, typeMeta.GroupVersionKind())
		}

		var findings []schema.GroupVersionKind

		g.Expect(actualGVKs).To(g.ContainElement(configMapGVK, &findings))
		g.Expect(findings).To(g.HaveLen(len(expectedConfigMapNames)))

		g.Expect(actualGVKs).To(g.HaveLen(len(expectedConfigMapNames)))
	})

	ginkgo.By("contains only expected Names", func() {
		var out bytes.Buffer
		g.Expect(main.GenerateManifests(kustomizeBuildYaml, &out)).To(g.Succeed())

		var actualNames []string
		for _, manifest := range separatorYaml.Split(out.String(), -1) {
			var objectMeta struct {
				metav1.ObjectMeta `json:"metadata"`
			}
			g.Expect(yaml.Unmarshal([]byte(manifest), &objectMeta)).To(g.Succeed())
			actualNames = append(actualNames, objectMeta.Name)
		}

		for _, expectedConfigMapName := range expectedConfigMapNames {
			g.Expect(actualNames).To(g.ContainElement(g.HavePrefix(expectedConfigMapName)))
		}

		g.Expect(actualNames).To(g.HaveLen(len(expectedConfigMapNames)))
	})
}
