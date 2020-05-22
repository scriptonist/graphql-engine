package acceptance

import (
	"io/ioutil"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("hasura root cmd", func() {
	var session *gexec.Session

	BeforeEach(func() {
		hasuraPath := buildHasura()
		cmd := exec.Command(hasuraPath)
		session = runCmd(cmd)
	})

	AfterEach(func() {
		gexec.CleanupBuildArtifacts()
	})

	It("prints help text to stdout and exits with a success code", func() {
		b, err := ioutil.ReadFile("testdata/rootCmdHelpText.golden")
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(session.Out.Contents()).To(BeEquivalentTo(string(b)))
	})
})
