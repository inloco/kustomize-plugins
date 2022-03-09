package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application"
	argov1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	separatorPanic  = ": "
	separatorYAML   = "---\n"
	yamlStatusField = "status"
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

type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ProjectSpec `json:"spec,omitempty"`
}

type ProjectSpec struct {
	AccessControl AppProjectAccessControl             `json:"accessControl,omitempty"`
	Destination   argov1alpha1.ApplicationDestination `json:"destination,omitempty"`
	AppProject    argov1alpha1.AppProject             `json:"appProjectTemplate,omitempty"`
	Applications  []argov1alpha1.Application          `json:"applicationTemplates,omitempty"`
}

type AppProjectAccessControl struct {
	ReadOnly  []string `json:"readOnly,omitempty"`
	ReadWrite []string `json:"readSync,omitempty"`
}

func main() {
	filePath := os.Args[1]

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Panic(filePath, separatorPanic, err)
	}

	var project Project
	if err := yaml.Unmarshal(data, &project); err != nil {
		log.Panic(filePath, separatorPanic, err)
	}

	argoAppProject, err := makeAppProject(&project)
	if err != nil {
		log.Panic(filePath, separatorPanic, err)
	}

	b, err := marshalAppProject(argoAppProject)
	if err != nil {
		log.Panic(filePath, separatorPanic, err)
	}

	if _, err := os.Stdout.Write(b); err != nil {
		log.Panic(filePath, separatorPanic, err)
	}

	for _, app := range project.Spec.Applications {
		if _, err := os.Stdout.Write([]byte(separatorYAML)); err != nil {
			log.Panic(filePath, separatorPanic, err)
		}

		b, err = makeApplication(argoAppProject, &app)
		if err != nil {
			log.Panic(filePath, separatorPanic, err)
		}

		if _, err := os.Stdout.Write(b); err != nil {
			log.Panic(filePath, separatorPanic, err)
		}
	}
}

func makeAppProject(project *Project) (*argov1alpha1.AppProject, error) {
	argoAppProject := project.Spec.AppProject

	argoAppProject.TypeMeta = metav1.TypeMeta{
		APIVersion: argov1alpha1.SchemeGroupVersion.String(),
		Kind:       application.AppProjectKind,
	}

	if argoAppProject.Name == "" {
		argoAppProject.Name = project.Name
	}

	argoAppProject.Spec.NamespaceResourceWhitelist = []metav1.GroupKind{
		metav1.GroupKind{
			Group: "*",
			Kind:  "*",
		},
	}

	argoAppProject.Spec.SourceRepos = []string{
		"*",
	}

	argoAppProject.Spec.Destinations = append(argoAppProject.Spec.Destinations, project.Spec.Destination)

	readOnlyProjectRole := makeProjectRole(readOnly, project, &argoAppProject)
	argoAppProject.Spec.Roles = append(argoAppProject.Spec.Roles, *readOnlyProjectRole)

	readSyncProjectRole := makeProjectRole(readSync, project, &argoAppProject)
	argoAppProject.Spec.Roles = append(argoAppProject.Spec.Roles, *readSyncProjectRole)

	return &argoAppProject, nil
}

func marshalAppProject(argoAppProject *argov1alpha1.AppProject) ([]byte, error) {
	return marshalYAMLWithoutStatusField(argoAppProject)
}

func makeProjectRole(accessLevel accessLevel, project *Project, appProject *argov1alpha1.AppProject) *argov1alpha1.ProjectRole {
	var groups []string
	switch accessLevel {
	case readOnly:
		groups = project.Spec.AccessControl.ReadOnly
	case readSync:
		groups = project.Spec.AccessControl.ReadWrite
	}

	return &argov1alpha1.ProjectRole{
		Name:     accessLevel.longName(),
		Policies: accessLevel.policies(appProject.Name),
		Groups:   groups,
	}
}

func makeApplication(argoAppProject *argov1alpha1.AppProject, app *argov1alpha1.Application) ([]byte, error) {
	app.TypeMeta = metav1.TypeMeta{
		APIVersion: argov1alpha1.SchemeGroupVersion.String(),
		Kind:       application.ApplicationKind,
	}

	if app.Name == "" {
		app.Name = argoAppProject.Name
	}

	app.Spec.Project = argoAppProject.Name
	app.Spec.Destination = argoAppProject.Spec.Destinations[0]

	return marshalYAMLWithoutStatusField(app)
}

func marshalYAMLWithoutStatusField(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var vm map[string]interface{}
	err = json.Unmarshal(b, &vm)
	if err != nil {
		return nil, err
	}

	delete(vm, yamlStatusField)

	return yaml.Marshal(vm)
}
