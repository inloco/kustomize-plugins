package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	yamlSeparator = "---\n"
)

type accessLevel int

const (
	readOnly accessLevel = iota
	readWrite
)

func (a accessLevel) longName() string {
	switch a {
	case readOnly:
		return "unnamespaced-ro"
	case readWrite:
		return "unnamespaced-rw"
	default:
		panic(fmt.Sprintf("unknown access level %d", a))
	}
}

func (a accessLevel) shortName() string {
	switch a {
	case readOnly:
		return "ro"
	case readWrite:
		return "rw"
	default:
		panic(fmt.Sprintf("unknown access level %d", a))
	}
}

type Namespace struct {
	metav1.TypeMeta `json:",inline"`
	AccessControl   NamespaceAccessControl `json:"accessControl,omitempty"`
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

	ro, err := makeClusterRoleBinding(readOnly, &namespace)
	if err != nil {
		return
	}
	yamls = append(yamls, ro)

	rw, err := makeClusterRoleBinding(readWrite, &namespace)
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

func makeClusterRoleBinding(accessLevel accessLevel, namespace *Namespace) ([]byte, error) {
	var names []string
	switch accessLevel {
	case readOnly:
		names = namespace.AccessControl.ReadOnly
	case readWrite:
		names = namespace.AccessControl.ReadWrite
	}

	return yaml.Marshal(rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       reflect.TypeOf(rbacv1.RoleBinding{}).Name(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: accessLevel.longName(),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     reflect.TypeOf(rbacv1.ClusterRole{}).Name(),
			Name:     accessLevel.longName(),
		},
		Subjects: makeSubjects(names),
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
