package main_test

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
)

func TestKustomizeBuild(t *testing.T) {
	g.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KustomizeBuild Suite")
}
