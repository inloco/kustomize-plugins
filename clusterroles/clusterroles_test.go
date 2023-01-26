package main_test

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"

	"github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"

	"github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	main "github.com/inloco/kustomize-generator-plugins/clusterroles"
)

var (
	yamlSeparator = regexp.MustCompile("\n---\n")

	clusterRoleGVK      = rbacv1.SchemeGroupVersion.WithKind(reflect.TypeOf(rbacv1.ClusterRole{}).Name())
	clusterRoleTypeMeta = metav1.TypeMeta{
		APIVersion: clusterRoleGVK.GroupVersion().String(),
		Kind:       clusterRoleGVK.Kind,
	}

	readOnlyVerbs = []string{
		main.VerbGet,
		main.VerbList,
		main.VerbWatch,
	}

	readWriteVerbs = []string{
		rbacv1.VerbAll,
	}
)

var apiResourceList = []*metav1.APIResourceList{
	{
		GroupVersion: main.CoreGroupName,
		APIResources: []metav1.APIResource{
			{
				Name:       "pods",
				Namespaced: true,
			},
			{
				Name:       "serviceaccounts",
				Namespaced: true,
			},
			{
				Name:       "services",
				Namespaced: true,
			},
			{
				Name:       "nodes",
				Namespaced: false,
			},
			{
				Name:       "namespaces",
				Namespaced: false,
			},
			{
				Name:       "persistentvolumes",
				Namespaced: false,
			},
		},
	},
	{
		GroupVersion: rbacv1.SchemeGroupVersion.String(),
		APIResources: []metav1.APIResource{
			{
				Name:       "clusterrolebindings",
				Namespaced: false,
			},
			{
				Name:       "clusterroles",
				Namespaced: false,
			},
			{
				Name:       "roles",
				Namespaced: true,
			},
		},
	},
	{
		GroupVersion: "",
		APIResources: []metav1.APIResource{
			{
				Name:       "",
				Namespaced: true,
			},
		},
	},
}

var _ = ginkgo.Describe("ClusterRoles", func() {
	ginkgo.It("contains only expected GKVs", func() {
		var out bytes.Buffer
		g.Expect(main.GenerateManifests(apiResourceList, &out)).To(g.BeNil())

		var actualGVKs []schema.GroupVersionKind
		for _, manifest := range yamlSeparator.Split(out.String(), -1) {
			var meta metav1.TypeMeta
			g.Expect(yaml.Unmarshal([]byte(manifest), &meta)).To(g.Succeed())
			actualGVKs = append(actualGVKs, meta.GroupVersionKind())
		}

		var findings []schema.GroupVersionKind

		g.Expect(actualGVKs).To(g.ContainElement(clusterRoleGVK, &findings))
		g.Expect(findings).To(g.HaveLen(4))

		g.Expect(actualGVKs).To(g.HaveLen(4))
	})

	ginkgo.It("contains expected ClusterRoles", func() {
		var out bytes.Buffer
		g.Expect(main.GenerateManifests(apiResourceList, &out)).To(g.BeNil())

		for _, manifest := range yamlSeparator.Split(out.String(), -1) {
			var clusterRole rbacv1.ClusterRole
			g.Expect(yaml.Unmarshal([]byte(manifest), &clusterRole)).To(g.Succeed())

			switch clusterRole.Name {
			case main.NamespacedReadOnlyRoleName:
				g.Expect(clusterRole).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"TypeMeta": g.Equal(clusterRoleTypeMeta),
					"ObjectMeta": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Name": g.Equal(main.NamespacedReadOnlyRoleName),
					}),
					"Rules": g.ContainElements([]rbacv1.PolicyRule{{
						APIGroups: []string{
							rbacv1.VerbAll,
						},
						Resources: []string{
							rbacv1.VerbAll,
						},
						Verbs: readOnlyVerbs,
					}}),
				}))
			case main.NamespacedReadWriteRoleName:
				g.Expect(clusterRole).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"TypeMeta": g.Equal(clusterRoleTypeMeta),
					"ObjectMeta": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Name": g.Equal(main.NamespacedReadWriteRoleName),
					}),
					"Rules": g.ConsistOf([]rbacv1.PolicyRule{{
						APIGroups: []string{
							rbacv1.VerbAll,
						},
						Resources: []string{
							rbacv1.VerbAll,
						},
						Verbs: []string{
							rbacv1.VerbAll,
						},
					}}),
				}))
			case main.UnnamespacedReadOnlyRoleName:
				g.Expect(clusterRole).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"TypeMeta": g.Equal(clusterRoleTypeMeta),
					"ObjectMeta": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Name": g.Equal(main.UnnamespacedReadOnlyRoleName),
					}),
					//"Rules": g.ContainElements([]rbacv1.PolicyRule{{
					//	APIGroups: []string{
					//		rbacv1.VerbAll,
					//	},
					//	Resources: []string{
					//		rbacv1.VerbAll,
					//	},
					//	Verbs: readOnlyVerbs,
					//}}),
				}))
				// only contains read verbs
				// contains all groups and resources expected
				// also check for size

			case main.UnnamespacedReadWriteRoleName:
				g.Expect(clusterRole).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"TypeMeta": g.Equal(clusterRoleTypeMeta),
					"ObjectMeta": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
						"Name": g.Equal(main.UnnamespacedReadWriteRoleName),
					}),
					//"Rules": g.ContainElements([]rbacv1.PolicyRule{{
					//	APIGroups: []string{
					//		main.CoreGroupName,
					//	},
					//	Resources: []string{
					//		rbacv1.VerbAll,
					//	},
					//	Verbs: []string{
					//		rbacv1.VerbAll,
					//	}}}),
				}))
			default:
				ginkgo.Fail(fmt.Sprintf("unexpected cluster role %s", clusterRole.Name))
			}
		}
	})
})
