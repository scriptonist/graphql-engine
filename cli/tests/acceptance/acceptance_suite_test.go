package acceptance

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var T *testing.T

var tempalateProjectV1Dir = "testdata/hasura_project_v1"
var tempalateProjectV2Dir = "testdata/hasura_project_v2"

var teardown func()

func TestAcceptance(t *testing.T) {
	T = t
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Suite")
}
