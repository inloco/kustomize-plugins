package main

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	readOnlyName  = "read-only"
	readWriteName = "read-write"
	yamlSeparator = "---\n"
)

type Namespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              corev1.NamespaceSpec   `json:"spec,omitempty"`
	AccessControl     NamespaceAccessControl `json:"accessControl,omitempty"`
}

type NamespaceAccessControl struct {
	ReadOnly  []string `json:"readOnly,omitempty"`
	ReadWrite []string `json:"readWrite,omitempty"`
}

func main() {
	filePath := os.Args[1]

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Panic(filePath, ": ", err)
	}

	var namespace Namespace
	if err := yaml.Unmarshal(data, &namespace); err != nil {
		log.Panic(filePath, ": ", err)
	}

	var yamls [][]byte

	ns, err := makeNamespace(&namespace)
	if err != nil {
		return
	}
	yamls = append(yamls, ns)

	ro, err := makeRoleBinding(readOnlyName, &namespace)
	if err != nil {
		return
	}
	yamls = append(yamls, ro)

	rw, err := makeRoleBinding(readWriteName, &namespace)
	if err != nil {
		return
	}
	yamls = append(yamls, rw)

	for _, y := range yamls {
		if _, err := os.Stdout.Write(y); err != nil {
			log.Panic(filePath, ": ", err)
		}

		if _, err := os.Stdout.Write([]byte(yamlSeparator)); err != nil {
			log.Panic(filePath, ": ", err)
		}
	}
}

func makeNamespace(namespace *Namespace) ([]byte, error) {
	return yaml.Marshal(corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       reflect.TypeOf(corev1.Namespace{}).Name(),
		},
		ObjectMeta: namespace.ObjectMeta,
		Spec:       namespace.Spec,
	})
}

func makeRoleBinding(name string, namespace *Namespace) ([]byte, error) {
	return yaml.Marshal(rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       reflect.TypeOf(rbacv1.RoleBinding{}).Name(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace.Name,
			Name:      name,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     reflect.TypeOf(rbacv1.ClusterRole{}).Name(),
			Name:     name,
		},
		Subjects: makeSubjects(namespace.AccessControl.ReadOnly),
	})
}

func makeSubjects(names []string) []rbacv1.Subject {
	var subjects []rbacv1.Subject
	for _, name := range names {
		subjects = append(subjects, rbacv1.Subject{
			APIGroup: rbacv1.GroupName,
			Kind:     rbacv1.GroupKind,
			Name:     name,
		})
	}

	return subjects
}
