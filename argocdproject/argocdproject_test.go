package main_test

import (
	"bytes"
	"fmt"
	"regexp"

	argov1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"

	main "github.com/inloco/iac-kustomize-plugins/argocdproject"
)

var (
	separatorYaml = regexp.MustCompile("\n---\n")
)

var _ = ginkgo.Describe("ArgoCDProject", func() {
	ginkgo.DescribeTable("", ArgoCDProject,
		ginkgo.Entry("with single application", main.ArgoCDProject{
			TypeMeta: metav1.TypeMeta{
				APIVersion: schema.GroupVersion{
					Group:   "incognia.com",
					Version: "v1alpha1",
				}.String(),
				Kind: "ArgoCDProject",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "github-checker",
			},
			Spec: main.ProjectSpec{
				AccessControl: main.AppProjectAccessControl{
					ReadOnly: []string{
						"sre:eng-2",
					},
					ReadSync: []string{
						"sre:eng-0",
						"sre:eng-1",
					},
				},
				Environment: "staging",
				ApplicationTemplates: []argov1alpha1.Application{
					argov1alpha1.Application{
						ObjectMeta: metav1.ObjectMeta{
							Name: "github-checker-app",
						},
						Spec: argov1alpha1.ApplicationSpec{
							Source: argov1alpha1.ApplicationSource{
								RepoURL: "https://github.com/inloco/github-checker.git",
							},
							Destination: argov1alpha1.ApplicationDestination{
								Name:      "arn:aws:eks:us:123456789876:cluster/Global-SRE",
								Namespace: "github-checker",
							},
						},
					},
				},
			},
		}),
		ginkgo.Entry("with multiple applications", main.ArgoCDProject{
			TypeMeta: metav1.TypeMeta{
				APIVersion: schema.GroupVersion{
					Group:   "incognia.com",
					Version: "v1alpha1",
				}.String(),
				Kind: "ArgoCDProject",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "github-checker",
			},
			Spec: main.ProjectSpec{
				AccessControl: main.AppProjectAccessControl{
					ReadSync: []string{
						"sre:eng-0",
					},
				},
				AppProject: argov1alpha1.AppProject{
					Spec: argov1alpha1.AppProjectSpec{
						Destinations: []argov1alpha1.ApplicationDestination{
							argov1alpha1.ApplicationDestination{
								Name:      "arn:aws:eks:us:123456789876:cluster/Global-SRE",
								Namespace: "github-checker",
							},
							argov1alpha1.ApplicationDestination{
								Name:      "arn:aws:eks:us:123456789876:cluster/Global-Product",
								Namespace: "another-checker",
							},
							argov1alpha1.ApplicationDestination{
								Name:      "arn:aws:eks:us:123456789876:cluster/Global-FortKnox",
								Namespace: "foreground-checker",
							},
						},
						ClusterResourceWhitelist: []metav1.GroupKind{
							metav1.GroupKind{
								Group: "*",
								Kind:  "*",
							},
						},
					},
				},
				ApplicationTemplates: []argov1alpha1.Application{
					argov1alpha1.Application{
						ObjectMeta: metav1.ObjectMeta{
							Name: "github-checker-app",
						},
						Spec: argov1alpha1.ApplicationSpec{
							Source: argov1alpha1.ApplicationSource{
								RepoURL:        "https://github.com/inloco/github-checker.git",
								Path:           "namespaces/example/environment-overlays/env/cluster-overlays/cluster",
								TargetRevision: "HEAD",
							},
							Destination: argov1alpha1.ApplicationDestination{
								Name:      "arn:aws:eks:us:123456789876:cluster/Global-SRE",
								Namespace: "github-checker",
							},
						},
					},
					argov1alpha1.Application{
						ObjectMeta: metav1.ObjectMeta{
							Name: "another-checker-app",
						},
						Spec: argov1alpha1.ApplicationSpec{
							Source: argov1alpha1.ApplicationSource{
								RepoURL:        "https://github.com/inloco/another-checker.git",
								Path:           "namespaces/example/environment-overlays/env/cluster-overlays/cluster",
								TargetRevision: "HEAD",
							},
							Destination: argov1alpha1.ApplicationDestination{
								Name:      "arn:aws:eks:us:123456789876:cluster/Global-Product",
								Namespace: "another-checker",
							},
						},
					},
				},
			},
		}),
	)
})

func ArgoCDProject(argoCDProject main.ArgoCDProject) {
	var argoCDProjectYaml []byte
	if data, err := yaml.Marshal(argoCDProject); g.Expect(err).To(g.BeNil()) {
		argoCDProjectYaml = data
	}

	ginkgo.By("contains only expected GKVs", func() {
		var out bytes.Buffer
		g.Expect(main.GenerateManifests(argoCDProjectYaml, &out)).To(g.Succeed())

		var actualGVKs []schema.GroupVersionKind
		for _, manifest := range separatorYaml.Split(out.String(), -1) {
			var meta metav1.TypeMeta
			g.Expect(yaml.Unmarshal([]byte(manifest), &meta)).To(g.Succeed())
			actualGVKs = append(actualGVKs, meta.GroupVersionKind())
		}

		var findings []schema.GroupVersionKind

		g.Expect(actualGVKs).To(g.ContainElement(argov1alpha1.AppProjectSchemaGroupVersionKind, &findings))
		g.Expect(findings).To(g.HaveLen(1))

		g.Expect(actualGVKs).To(g.ContainElement(argov1alpha1.ApplicationSchemaGroupVersionKind, &findings))
		g.Expect(findings).To(g.HaveLen(len(argoCDProject.Spec.ApplicationTemplates)))

		g.Expect(actualGVKs).To(g.HaveLen(1 + len(argoCDProject.Spec.ApplicationTemplates)))
	})

	ginkgo.By("contains expected AppProject", func() {
		var out bytes.Buffer
		g.Expect(main.GenerateManifests(argoCDProjectYaml, &out)).To(g.Succeed())

		var appProject argov1alpha1.AppProject
		for _, manifest := range separatorYaml.Split(out.String(), -1) {
			var meta metav1.TypeMeta
			g.Expect(yaml.Unmarshal([]byte(manifest), &meta)).To(g.Succeed())

			if meta.GroupVersionKind() == argov1alpha1.AppProjectSchemaGroupVersionKind {
				g.Expect(yaml.Unmarshal([]byte(manifest), &appProject)).To(g.Succeed())
				break
			}
		}

		destinations := argoCDProject.Spec.AppProject.Spec.Destinations
		if destinations == nil {
			destinationMap := make(map[string]argov1alpha1.ApplicationDestination)
			for _, applicationTemplate := range argoCDProject.Spec.ApplicationTemplates {
				destination := applicationTemplate.Spec.Destination
				destinationMap[destination.String()] = destination
			}

			destinations = make([]argov1alpha1.ApplicationDestination, 0, len(destinationMap))
			for _, destination := range destinationMap {
				destinations = append(destinations, destination)
			}
		}
		specDestinationsMatcher := g.And(
			g.ContainElements(destinations),
			g.HaveLen(len(destinations)),
		)

		g.Expect(appProject).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
			"ObjectMeta": g.Equal(argoCDProject.ObjectMeta),
			"Spec": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"SourceRepos": g.Equal([]string{
					"*",
				}),
				"Destinations":             specDestinationsMatcher,
				"ClusterResourceWhitelist": g.Equal(argoCDProject.Spec.AppProject.Spec.ClusterResourceWhitelist),
				"NamespaceResourceWhitelist": g.Equal([]metav1.GroupKind{{
					Group: "*",
					Kind:  "*",
				}}),
				"Roles": gstruct.MatchAllElements(func(e interface{}) string {
					return e.(argov1alpha1.ProjectRole).Name
				}, gstruct.Elements{
					main.ReadOnly.String(): gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Groups":   g.ContainElements(argoCDProject.Spec.AccessControl.ReadOnly),
						"Policies": g.ContainElements(main.ReadOnly.Policies(argoCDProject.Name, argoCDProject.Spec.Environment)),
					}),
					main.ReadSync.String(): gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Groups":   g.ContainElements(argoCDProject.Spec.AccessControl.ReadSync),
						"Policies": g.ContainElements(main.ReadSync.Policies(argoCDProject.Name, argoCDProject.Spec.Environment)),
					}),
				}),
			}),
		}))
	})

	ginkgo.By("contains expected Applications", func() {
		var out bytes.Buffer
		g.Expect(main.GenerateManifests(argoCDProjectYaml, &out)).To(g.Succeed())

		for _, manifest := range separatorYaml.Split(out.String(), -1) {
			var meta metav1.TypeMeta
			g.Expect(yaml.Unmarshal([]byte(manifest), &meta)).To(g.Succeed())

			if meta.GroupVersionKind() != argov1alpha1.ApplicationSchemaGroupVersionKind {
				continue
			}

			var app argov1alpha1.Application
			g.Expect(yaml.Unmarshal([]byte(manifest), &app)).To(g.Succeed())

			var argoCdProjectApp argov1alpha1.Application
			for _, appTemplate := range argoCDProject.Spec.ApplicationTemplates {
				if app.Name == appTemplate.Name {
					argoCdProjectApp = appTemplate
					break
				}
			}

			specSourcePathMatcher := g.Equal(argoCdProjectApp.Spec.Source.Path)
			specSourceTargetRevisionMatcher := g.Equal(argoCdProjectApp.Spec.Source.TargetRevision)

			if argoCDProject.Spec.Environment != "" {
				specSourcePathMatcher = g.Equal(fmt.Sprintf("./k8s/overlays/%s", argoCDProject.Spec.Environment))
				specSourceTargetRevisionMatcher = g.Equal(fmt.Sprintf("env-%s", argoCDProject.Spec.Environment))
			}

			g.Expect(app).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"ObjectMeta": g.Equal(argoCdProjectApp.ObjectMeta),
				"Spec": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Project": g.Equal(argoCDProject.Name),
					"Source": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"RepoURL":        g.Equal(argoCdProjectApp.Spec.Source.RepoURL),
						"Path":           specSourcePathMatcher,
						"TargetRevision": specSourceTargetRevisionMatcher,
					}),
					"Destination": g.Equal(argoCdProjectApp.Spec.Destination),
				}),
			}))
		}

		ginkgo.By("manifests contain nil status field", func() {
			var out bytes.Buffer
			g.Expect(main.GenerateManifests(argoCDProjectYaml, &out)).To(g.Succeed())

			for _, manifest := range separatorYaml.Split(out.String(), -1) {
				var resource map[string]interface{}
				g.Expect(yaml.Unmarshal([]byte(manifest), &resource)).To(g.Succeed())
				g.Expect(resource["status"]).To(g.BeNil())
			}
		})
	})
}
