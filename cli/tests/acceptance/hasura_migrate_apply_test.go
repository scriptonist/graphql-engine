package acceptance

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var runMigrateApplyWithGlobalFlags = func(globalFlags ...string) func() {
	return func() {
		Describe("hasura migrate apply", func() {
			var hasuraBinary string
			var workingDir string
			var cmd *exec.Cmd
			var teardown func()
			var baseArgs = []string{"hasuraBinary", "migrate", "apply"}

			BeforeEach(func() {
				hasuraBinary = buildHasura()
				networkName := "haura-migrate-apply-" + randString()

				// create network
				removeNetwork, err := newTestNetwork(networkName)
				Expect(err).NotTo(HaveOccurred())

				// create postgres container
				postgresUrl, removePostgresC, err := newPostgresContainer(networkName)
				Expect(err).NotTo(HaveOccurred())

				// create hge container
				hgeUrl, removeHgeC, err := newHgeContainer(nil, networkName, postgresUrl)
				Expect(err).NotTo(HaveOccurred())

				tempProjectDir := duplicateProjectDirectory(tempalateProjectV1Dir, hgeUrl)
				workingDir = fmt.Sprintf("%s/%s", tempProjectDir, "testdata")
				teardown = func() {
					removeHgeC()
					removePostgresC()
					removeNetwork()
				}
			})

			AfterEach(func() {
				teardown()
			})

			It("migrates up 2 steps", func() {
				cmd = exec.Command(hasuraBinary)
				args := append(baseArgs, []string{"--up", "2"}...)
				cmd.Dir = workingDir
				cmd.Args = args
				session := runCmd(cmd)

				Eventually(session, "10s").Should(gexec.Exit(0))
				Eventually(session, "10s").Should(gbytes.Say("Applying migrations..."))
				Eventually(session, "10s").Should(gbytes.Say("migrations applied"))
			})

			It("migrates down 2 steps", func() {})
			It("apply up of a particular version", func() {})
			It("applies down migration of a particular version", func() {})
		})
	}

}

var _ = Describe("applying migrations", func() {
	Context("no global flags", runMigrateApplyWithGlobalFlags())
	//Context("--config-file of version V1 is provided", runMigrateApplyWithGlobalFlags("--config-file", "testdata/hasura_project_v1/config.yaml"))
	//Context("--config-file of version V2 is provided", runMigrateApplyWithGlobalFlags("--config-file", "testdata/config-v2.yaml"))
})
