package main

import (
	"io"
	"log"
	"os"
	"text/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	panicSeparator = ": "

	tmplName   = "_"
	tmplPrefix = "([{"
	tmplSuffix = "}])"
	tmplOption = "missingkey=error"
)

type Template struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Data              map[string]interface{} `json:"data,omitempty"`
}

func main() {
	filePath := os.Args[1]

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Panic(filePath, panicSeparator, err)
	}

	if err := GenerateManifests(data, os.Stdout); err != nil {
		log.Panic(filePath, panicSeparator, err)
	}
}

func GenerateManifests(data []byte, out io.Writer) error {
	var templateTransform Template
	if err := yaml.Unmarshal(data, &templateTransform); err != nil {
		return err
	}

	tmpl, err := template.New(tmplName).
		Delims(tmplPrefix, tmplSuffix).
		Option(tmplOption).
		ParseFS(virtualFS{}, tmplName)
	if err != nil {
		return err
	}

	return tmpl.Execute(out, templateTransform.Data)
}
