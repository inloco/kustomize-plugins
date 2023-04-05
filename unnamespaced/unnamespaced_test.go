package main_test

import (
	"bytes"
	"reflect"
	"regexp"

	"github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"

	main "github.com/inloco/kustomize-plugins/unnamespaced"
)

var (
	separatorYaml = regexp.MustCompile("\n---\n")

	clusterRoleBindingGVK = rbacv1.SchemeGroupVersion.WithKind(reflect.TypeOf(rbacv1.ClusterRoleBinding{}).Name())
)

var _ = ginkgo.Describe("Unnamespace", func() {
	ginkgo.DescribeTable("", Unnamespaced,
		ginkgo.Entry("with complete access control", main.Unnamespaced{
			TypeMeta: metav1.TypeMeta{
				APIVersion: schema.GroupVersion{
					Group:   "incognia.com",
					Version: "v1alpha1",
				}.String(),
				Kind: "Unnamespace",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "_",
			},
			AccessControl: main.UnnamespacedAccessControl{
				ReadOnly: []string{
					"sre:eng-2",
				},
				ReadWrite: []string{
					"sre:eng-0",
					"sre:eng-1",
				},
			},
		}),
		ginkgo.Entry("with partial access control", main.Unnamespaced{
			TypeMeta: metav1.TypeMeta{
				APIVersion: schema.GroupVersion{
					Group:   "incognia.com",
					Version: "v1alpha1",
				}.String(),
				Kind: "Unnamespace",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "_",
			},
			AccessControl: main.UnnamespacedAccessControl{
				ReadWrite: []string{
					"sre:eng-0",
					"sre:eng-1",
				},
			},
		}),
		ginkgo.Entry("with empty access control", main.Unnamespaced{
			TypeMeta: metav1.TypeMeta{
				APIVersion: schema.GroupVersion{
					Group:   "incognia.com",
					Version: "v1alpha1",
				}.String(),
				Kind: "Unnamespace",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "_",
			},
		}),
	)
})

func Unnamespaced(unnamespaced main.Unnamespaced) {
	var unnamespaceYaml []byte
	if data, err := yaml.Marshal(unnamespaced); g.Expect(err).To(g.BeNil()) {
		unnamespaceYaml = data
	}

	ginkgo.By("contains only expected GKVs", func() {
		var out bytes.Buffer
		g.Expect(main.GenerateManifests(unnamespaceYaml, &out)).To(g.Succeed())

		var actualGVKs []schema.GroupVersionKind
		for _, resource := range separatorYaml.Split(out.String(), -1) {
			var meta metav1.TypeMeta
			g.Expect(yaml.Unmarshal([]byte(resource), &meta)).To(g.Succeed())
			actualGVKs = append(actualGVKs, meta.GroupVersionKind())
		}

		var findings []schema.GroupVersionKind

		g.Expect(actualGVKs).To(g.ContainElement(clusterRoleBindingGVK, &findings))
		g.Expect(findings).To(g.HaveLen(2))

		g.Expect(actualGVKs).To(g.HaveLen(2))
	})

	ginkgo.By("contains expected ClusterRoleBindings", func() {
		var out bytes.Buffer
		g.Expect(main.GenerateManifests(unnamespaceYaml, &out)).To(g.Succeed())

		for _, manifest := range separatorYaml.Split(out.String(), -1) {
			var clusterRoleBinding rbacv1.ClusterRoleBinding
			g.Expect(yaml.Unmarshal([]byte(manifest), &clusterRoleBinding)).To(g.Succeed())

			accessLevel := main.AccessLevelFromLongName(clusterRoleBinding.Name)

			var names []string
			switch accessLevel {
			case main.ReadOnly:
				names = unnamespaced.AccessControl.ReadOnly
			case main.ReadWrite:
				names = unnamespaced.AccessControl.ReadWrite
			default:
				ginkgo.Fail("unknown access level")
			}

			var subjects []rbacv1.Subject
			if len(names) > 0 {
				subjects = make([]rbacv1.Subject, 0, len(names))
				for _, name := range names {
					subjects = append(subjects, rbacv1.Subject{
						APIGroup: rbacv1.GroupName,
						Kind:     rbacv1.GroupKind,
						Name:     name,
					})
				}
			}

			g.Expect(clusterRoleBinding).To(gstruct.MatchAllFields(gstruct.Fields{
				"TypeMeta": g.Equal(metav1.TypeMeta{
					APIVersion: clusterRoleBindingGVK.GroupVersion().String(),
					Kind:       clusterRoleBindingGVK.Kind,
				}),
				"ObjectMeta": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Name": g.Or(
						g.Equal(main.ReadOnly.LongName()),
						g.Equal(main.ReadWrite.LongName()),
					),
				}),
				"RoleRef": g.Equal(rbacv1.RoleRef{
					APIGroup: rbacv1.GroupName,
					Kind:     reflect.TypeOf(rbacv1.ClusterRole{}).Name(),
					Name:     clusterRoleBinding.Name,
				}),
				"Subjects": g.Equal(subjects),
			}))
		}
	})
}
