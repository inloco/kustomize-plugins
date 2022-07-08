package main_test

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"

	"github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"

	"github.com/inloco/iac-kustomize-plugins/namespace"
)

var (
	separatorYaml  = regexp.MustCompile("\n---\n")
	namespaceGVK   = corev1.SchemeGroupVersion.WithKind(reflect.TypeOf(corev1.Namespace{}).Name())
	roleBindingGVK = rbacv1.SchemeGroupVersion.WithKind(reflect.TypeOf(rbacv1.RoleBinding{}).Name())
)

var _ = ginkgo.Describe("Namespace", func() {
	ginkgo.DescribeTable("entries", Namespace,
		ginkgo.Entry("with access control", main.Namespace{
			TypeMeta: metav1.TypeMeta{
				APIVersion: schema.GroupVersion{
					Group:   "incognia.com",
					Version: "v1alpha1",
				}.String(),
				Kind: "Namespace",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "example",
			},
			AccessControl: main.NamespaceAccessControl{
				ReadOnly: []string{
					"sre:eng-2",
				},
				ReadWrite: []string{
					"sre:eng-0",
					"sre:eng-1",
				},
			},
			Spec: corev1.NamespaceSpec{
				Finalizers: []corev1.FinalizerName{
					"finalizer",
				},
			},
		}),
	)
})

func Namespace(incogniaNamespace main.Namespace) {
	var incogniaNamespaceYaml []byte
	if data, err := yaml.Marshal(incogniaNamespace); g.Expect(err).To(g.BeNil()) {
		incogniaNamespaceYaml = data
	}

	ginkgo.By("contains only expected GKVs", func() {
		var out bytes.Buffer
		g.Expect(main.GenerateManifests(incogniaNamespaceYaml, &out)).To(g.Succeed())

		var actualGVKs []schema.GroupVersionKind
		for _, resource := range separatorYaml.Split(out.String(), -1) {
			var meta metav1.TypeMeta
			g.Expect(yaml.Unmarshal([]byte(resource), &meta)).To(g.Succeed())
			actualGVKs = append(actualGVKs, meta.GroupVersionKind())
		}

		var findings []schema.GroupVersionKind

		g.Expect(actualGVKs).To(g.ContainElement(namespaceGVK, &findings))
		g.Expect(findings).To(g.HaveLen(1))

		g.Expect(actualGVKs).To(g.ContainElement(roleBindingGVK, &findings))
		g.Expect(findings).To(g.HaveLen(2))

		g.Expect(actualGVKs).To(g.HaveLen(3))
	})

	ginkgo.By("contains expected Namespace", func() {
		var out bytes.Buffer
		g.Expect(main.GenerateManifests(incogniaNamespaceYaml, &out)).To(g.Succeed())

		var namespace corev1.Namespace
		for _, manifest := range separatorYaml.Split(out.String(), -1) {
			var meta metav1.TypeMeta
			g.Expect(yaml.Unmarshal([]byte(manifest), &meta)).To(g.Succeed())

			if meta.GroupVersionKind() == namespaceGVK {
				g.Expect(yaml.Unmarshal([]byte(manifest), &namespace)).To(g.Succeed())
				break
			}
		}

		g.Expect(namespace).To(gstruct.MatchAllFields(gstruct.Fields{
			"TypeMeta": g.Equal(metav1.TypeMeta{
				APIVersion: namespaceGVK.GroupVersion().String(),
				Kind:       namespaceGVK.Kind,
			}),
			"ObjectMeta": g.Equal(incogniaNamespace.ObjectMeta),
			"Spec":       g.Equal(incogniaNamespace.Spec),
			"Status":     g.Equal(incogniaNamespace.Status),
		}))
	})

	ginkgo.By("contains expected RoleBindings", func() {
		var out bytes.Buffer
		g.Expect(main.GenerateManifests(incogniaNamespaceYaml, &out)).To(g.Succeed())

		for _, manifest := range separatorYaml.Split(out.String(), -1) {
			var meta metav1.TypeMeta
			g.Expect(yaml.Unmarshal([]byte(manifest), &meta)).To(g.Succeed())

			if meta.GroupVersionKind() != roleBindingGVK {
				continue
			}

			var roleBinding rbacv1.RoleBinding
			g.Expect(yaml.Unmarshal([]byte(manifest), &roleBinding)).To(g.Succeed())

			accessLevel := main.AccessLevelFromLongName(roleBinding.Name)

			var names []string
			switch accessLevel {
			case main.ReadOnly:
				names = incogniaNamespace.AccessControl.ReadOnly
			case main.ReadWrite:
				names = incogniaNamespace.AccessControl.ReadWrite
			default:
				ginkgo.Fail("unknown access level")
			}
			names = append(names, fmt.Sprintf("%s:%s", roleBinding.Namespace, accessLevel.ShortName()))

			subjects := make([]rbacv1.Subject, 0, len(names))
			for _, name := range names {
				subjects = append(subjects, rbacv1.Subject{
					APIGroup: rbacv1.GroupName,
					Kind:     rbacv1.GroupKind,
					Name:     name,
				})
			}

			g.Expect(roleBinding).To(gstruct.MatchAllFields(gstruct.Fields{
				"TypeMeta": g.Equal(metav1.TypeMeta{
					APIVersion: roleBindingGVK.GroupVersion().String(),
					Kind:       roleBindingGVK.Kind,
				}),
				"ObjectMeta": gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
					"Name": g.Or(
						g.Equal(main.ReadOnly.LongName()),
						g.Equal(main.ReadWrite.LongName()),
					),
					"Namespace": g.Equal(incogniaNamespace.Name),
				}),
				"RoleRef": g.Equal(rbacv1.RoleRef{
					APIGroup: rbacv1.GroupName,
					Kind:     reflect.TypeOf(rbacv1.ClusterRole{}).Name(),
					Name:     roleBinding.Name,
				}),
				"Subjects": g.And(
					g.ContainElements(subjects),
					g.HaveLen(len(subjects)),
				),
			}))
		}
	})
}
