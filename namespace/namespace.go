package main

import (
	"fmt"
	"io"
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
	yamlSeparator = "---\n"
)

type AccessLevel int

const (
	ReadOnly AccessLevel = iota
	ReadWrite
)

func (a AccessLevel) LongName() string {
	switch a {
	case ReadOnly:
		return "namespaced-ro"
	case ReadWrite:
		return "namespaced-rw"
	default:
		panic(fmt.Sprintf("unknown access level %d", a))
	}
}

func (a AccessLevel) ShortName() string {
	switch a {
	case ReadOnly:
		return "ro"
	case ReadWrite:
		return "rw"
	default:
		panic(fmt.Sprintf("unknown access level %d", a))
	}
}

func AccessLevelFromLongName(s string) AccessLevel {
	switch s {
	case ReadOnly.LongName():
		return ReadOnly
	case ReadWrite.LongName():
		return ReadWrite
	default:
		panic(fmt.Sprintf("unknown access level %s", s))
	}
}

type Namespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	AccessControl     NamespaceAccessControl `json:"accessControl,omitempty"`
	Spec              corev1.NamespaceSpec   `json:"spec,omitempty"`
	Status            corev1.NamespaceStatus `json:"status,omitempty"`
}

type NamespaceAccessControl struct {
	ReadOnly  []string `json:"ReadOnly,omitempty"`
	ReadWrite []string `json:"ReadWrite,omitempty"`
}

func main() {
	filePath := os.Args[1]

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Panic(filePath, ": ", err)
	}

	if err := GenerateManifests(data, os.Stdout); err != nil {
		log.Panic(filePath, ": ", err)
	}
}

func GenerateManifests(data []byte, out io.Writer) error {
	var namespace Namespace
	if err := yaml.Unmarshal(data, &namespace); err != nil {
		return err
	}

	var manifests [][]byte

	ns, err := makeNamespace(&namespace)
	if err != nil {
		return err
	}
	manifests = append(manifests, ns)

	readOnlyRoleBinding, err := makeRoleBinding(ReadOnly, &namespace)
	if err != nil {
		return err
	}
	manifests = append(manifests, readOnlyRoleBinding)

	readWriteRoleBinding, err := makeRoleBinding(ReadWrite, &namespace)
	if err != nil {
		return err
	}
	manifests = append(manifests, readWriteRoleBinding)

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

func makeNamespace(namespace *Namespace) ([]byte, error) {
	ns := corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       reflect.TypeOf(corev1.Namespace{}).Name(),
		},
		ObjectMeta: namespace.ObjectMeta,
		Spec:       namespace.Spec,
		Status:     namespace.Status,
	}

	return yaml.Marshal(ns)
}

func makeRoleBinding(accessLevel AccessLevel, namespace *Namespace) ([]byte, error) {
	var names []string
	switch accessLevel {
	case ReadOnly:
		names = namespace.AccessControl.ReadOnly
	case ReadWrite:
		names = namespace.AccessControl.ReadWrite
	}
	names = append(names, fmt.Sprintf("%s:%s", namespace.GetName(), accessLevel.ShortName()))

	roleBinding := rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       reflect.TypeOf(rbacv1.RoleBinding{}).Name(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace.GetName(),
			Name:      accessLevel.LongName(),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     reflect.TypeOf(rbacv1.ClusterRole{}).Name(),
			Name:     accessLevel.LongName(),
		},
		Subjects: makeSubjects(names),
	}

	return yaml.Marshal(roleBinding)
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
