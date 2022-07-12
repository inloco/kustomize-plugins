package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	panicSeparator = ": "
	yamlSeparator  = "---\n"
)

type AccessLevel int

const (
	ReadOnly AccessLevel = iota
	ReadWrite
)

func (a AccessLevel) LongName() string {
	switch a {
	case ReadOnly:
		return "unnamespaced-ro"
	case ReadWrite:
		return "unnamespaced-rw"
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

type Unnamespaced struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	AccessControl     UnnamespacedAccessControl `json:"accessControl,omitempty"`
}

type UnnamespacedAccessControl struct {
	ReadOnly  []string `json:"ReadOnly,omitempty"`
	ReadWrite []string `json:"ReadWrite,omitempty"`
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
	var unnamespaced Unnamespaced
	if err := yaml.Unmarshal(data, &unnamespaced); err != nil {
		return err
	}

	manifests, err := makeManifests(&unnamespaced)
	if err != nil {
		return err
	}

	for _, y := range manifests {
		if _, err := out.Write([]byte(yamlSeparator)); err != nil {
			return err
		}

		if _, err := out.Write(y); err != nil {
			return err
		}
	}

	return nil
}

func makeManifests(unnamespaced *Unnamespaced) ([][]byte, error) {
	var manifests [][]byte

	readOnlyClusterRoleBinding, err := makeClusterRoleBinding(ReadOnly, unnamespaced)
	if err != nil {
		return nil, err
	}
	manifests = append(manifests, readOnlyClusterRoleBinding)

	readWriteClusterRoleBinding, err := makeClusterRoleBinding(ReadWrite, unnamespaced)
	if err != nil {
		return nil, err
	}
	manifests = append(manifests, readWriteClusterRoleBinding)

	return manifests, nil
}

func makeClusterRoleBinding(accessLevel AccessLevel, unnamespaced *Unnamespaced) ([]byte, error) {
	var names []string
	switch accessLevel {
	case ReadOnly:
		names = unnamespaced.AccessControl.ReadOnly
	case ReadWrite:
		names = unnamespaced.AccessControl.ReadWrite
	}

	objectMeta := unnamespaced.ObjectMeta
	objectMeta.Name = accessLevel.LongName()

	clusterRoleBinding := rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       reflect.TypeOf(rbacv1.ClusterRoleBinding{}).Name(),
		},
		ObjectMeta: objectMeta,
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     reflect.TypeOf(rbacv1.ClusterRole{}).Name(),
			Name:     accessLevel.LongName(),
		},
		Subjects: makeSubjects(names),
	}

	return yaml.Marshal(clusterRoleBinding)
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
