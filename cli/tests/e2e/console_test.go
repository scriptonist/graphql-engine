package e2e

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/hasura/graphql-engine/cli/tests/e2e/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("console command", func() {
	Context("v1 project", func() {
		var teardown func()
		BeforeEach(func() {
			projectDir := helpers.RandDirName()
			helpers.RunCommandAndSucceed("init", projectDir, "--version", "1")
			err := os.Chdir(projectDir)
			Expect(err).To(BeNil())
			session := helpers.Hasura("console", "--no-browser")
			want := `.*console running at: http://localhost:9695/.*`
			Eventually(session, 5).Should(Say(want))

			teardown = func() {
				session.Signal(os.Interrupt)
				os.RemoveAll(projectDir)
				os.Chdir(os.TempDir())
			}
		})
		AfterEach(func() {
			teardown()
		})

		When("api requests are send", func() {
			It("generates metadata and migration files", func() {
				checkRequiredServersAreStarted()
				sendAPIRequestsToCreateMetadataAndMigrations()
			})
		})
	})
	Context("v2 project", func() {
		var teardown func()
		BeforeEach(func() {
			projectDir := helpers.RandDirName()
			fmt.Fprintln(GinkgoWriter, "creating this", projectDir)
			helpers.RunCommandAndSucceed("init", projectDir)
			err := os.Chdir(projectDir)
			Expect(err).To(BeNil())
			session := helpers.Hasura("console", "--no-browser")
			want := `.*console running at: http://localhost:9695/.*`
			Eventually(session, 5).Should(Say(want))

			teardown = func() {
				session.Terminate()
				os.RemoveAll(projectDir)
				os.Chdir(os.TempDir())
			}
		})
		AfterEach(func() {
			teardown()
		})
		When("api requests are send", func() {
			It("generates metadata and migration files", func() {
				checkRequiredServersAreStarted()
				sendAPIRequestsToCreateMetadataAndMigrations()
			})
		})
	})
})

func checkRequiredServersAreStarted() {
	resp, err := http.Get("http://localhost:9695/console")
	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	resp, err = http.Get("http://localhost:9693/apis/migrate")
	Expect(err).To(BeNil())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	b, err := ioutil.ReadAll(resp.Body)
	Expect(b).Should(MatchJSON(`{"migrations":[],"status":{}}`))
	fmt.Fprint(GinkgoWriter, string(b))
	defer resp.Body.Close()
}

func sendAPIRequestsToCreateMetadataAndMigrations() {
}
