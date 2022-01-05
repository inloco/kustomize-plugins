package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application"
	argov1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type accessLevel int

const (
	readOnly accessLevel = iota
	readSync
)

func (a accessLevel) longName() string {
	switch a {
	case readOnly:
		return "read-only"
	case readSync:
		return "read-sync"
	default:
		panic(fmt.Sprintf("unknown access level %d", a))
	}
}

func (a accessLevel) policies(appProjectName string) []string {
	switch a {
	case readOnly:
		return []string{
			fmt.Sprintf("p, proj:%s:read-only, *, get, %s/*, allow", appProjectName, appProjectName),
		}
	case readSync:
		return []string{
			fmt.Sprintf("p, proj:%s:read-sync, applications, sync, %s/*, allow", appProjectName, appProjectName),
			fmt.Sprintf("g, proj:%s:read-sync, proj:%s:read-only", appProjectName, appProjectName),
		}
	default:
		panic(fmt.Sprintf("unknown access level %d", a))
	}
}

type AppProject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              argov1alpha1.AppProjectSpec   `json:"spec,omitempty"`
	Status            argov1alpha1.AppProjectStatus `json:"status,omitempty"`
	AccessControl     AppProjectAccessControl       `json:"accessControl,omitempty"`
}

type AppProjectAccessControl struct {
	ReadOnly  []string `json:"readOnly,omitempty"`
	ReadWrite []string `json:"readSync,omitempty"`
}

func main() {
	filePath := os.Args[1]

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Panic(filePath, ": ", err)
	}

	var appProject AppProject
	if err := yaml.Unmarshal(data, &appProject); err != nil {
		log.Panic(filePath, ": ", err)
	}

	b, err := makeAppProject(&appProject)
	if err != nil {
		return
	}

	if _, err := os.Stdout.Write(b); err != nil {
		log.Panic(filePath, ": ", err)
	}
}

func makeAppProject(appProject *AppProject) ([]byte, error) {
	readOnlyProjectRole := makeProjectRole(readOnly, appProject)
	appProject.Spec.Roles = append(appProject.Spec.Roles, *readOnlyProjectRole)

	readSyncProjectRole := makeProjectRole(readSync, appProject)
	appProject.Spec.Roles = append(appProject.Spec.Roles, *readSyncProjectRole)

	return yaml.Marshal(argov1alpha1.AppProject{
		TypeMeta: metav1.TypeMeta{
			APIVersion: argov1alpha1.SchemeGroupVersion.String(),
			Kind:       application.AppProjectKind,
		},
		ObjectMeta: appProject.ObjectMeta,
		Spec:       appProject.Spec,
		Status:     appProject.Status,
	})
}

func makeProjectRole(accessLevel accessLevel, appProject *AppProject) *argov1alpha1.ProjectRole {
	var groups []string
	switch accessLevel {
	case readOnly:
		groups = appProject.AccessControl.ReadOnly
	case readSync:
		groups = appProject.AccessControl.ReadWrite
	}
	return &argov1alpha1.ProjectRole{
		Name:     accessLevel.longName(),
		Policies: accessLevel.policies(appProject.Name),
		Groups:   groups,
	}
}
