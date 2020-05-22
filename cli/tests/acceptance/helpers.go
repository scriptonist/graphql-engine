package acceptance

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/hasura/graphql-engine/cli/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/testcontainers/testcontainers-go"
	"gopkg.in/yaml.v3"
)

func buildHasura() string {
	hasuraPath, err := gexec.Build("github.com/hasura/graphql-engine/cli/cmd/hasura")
	Expect(err).NotTo(HaveOccurred())

	return hasuraPath
}

func runCmd(cmd *exec.Cmd) *gexec.Session {
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return session
}

func newPostgresContainer(networkName string) (string, func(), error) {
	const postgresUser = "postgres"
	const postgresPassword = "password"
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:12",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": postgresPassword,
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
		Networks:   []string{networkName},
	}
	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return "", nil, err
	}
	ip, err := postgresC.Host(ctx)
	if err != nil {
		return "", nil, err
	}
	log.Println(ip)
	port, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		return "", nil, err
	}
	url := fmt.Sprintf("postgres://%s:%s@%s:%s", postgresUser, postgresPassword, postgresC.GetContainerID(), port.Port())
	fmt.Println(url)

	removePostgresC := func() {
		err := postgresC.Terminate(ctx)
		Expect(err).NotTo(HaveOccurred())
	}
	return url, removePostgresC, nil
}

func newHgeContainer(additionalEnvVars map[string]string, networkName string, postgresUrl string) (string, func(), error) {
	envVars := map[string]string{
		"HASURA_GRAPHQL_DATABASE_URL":       fmt.Sprintf("%s/%s", postgresUrl, "postgres"),
		"HASURA_GRAPHQL_ENABLE_CONSOLE":     "true",
		"HASURA_GRAPHQL_ENABLED_LOG_TYPES":  "startup, http-log, webhook-log, websocket-log, query-log",
		"HASURA_GRAPHQL_CONSOLE_ASSETS_DIR": "/srv/console-assets",
		"HASURA_GRAPHQL_LOG_LEVEL":          "debug",
	}
	for k, v := range additionalEnvVars {
		envVars[k] = v
	}
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "hasura/graphql-engine:v1.2.1",
		ExposedPorts: []string{"8080/tcp"},
		Networks:     []string{networkName},
		Env:          envVars,
		WaitingFor:   wait.ForHTTP("/v1/version").WithPort("8080"),
	}
	hgeC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return "", nil, err
	}
	//ip, err := hgeC.Host(ctx)
	//if err != nil {
	//	return "", nil, err
	//}
	port, err := hgeC.MappedPort(ctx, "8080")
	if err != nil {
		return "", nil, err
	}
	url := fmt.Sprintf("http://%s:%s", "gateway.docker.internal", port.Port())

	removeHgeC := func() {
		err := hgeC.Terminate(ctx)
		Expect(err).NotTo(HaveOccurred())
	}
	return url, removeHgeC, nil
}

func duplicateProjectDirectory(from string, hgeEndpoint string) string {
	// create a temp dir
	tempDir, err := ioutil.TempDir(os.TempDir(), "*-cli-test")
	Expect(err).NotTo(HaveOccurred())

	// copy project template to tempdir
	err = util.CopyDir(from, fmt.Sprintf("%s/%s", tempDir, "testdata"))
	Expect(err).NotTo(HaveOccurred())

	// edit config file and change endpoint
	// read
	configFilePath := fmt.Sprintf("%s/%s/%s", tempDir, "testdata", "config.yaml")
	b, err := ioutil.ReadFile(configFilePath)
	Expect(err).NotTo(HaveOccurred())
	var c map[string]string
	err = yaml.Unmarshal(b, &c)
	Expect(err).NotTo(HaveOccurred())

	// edit
	Expect(c).To(HaveKey("endpoint"))
	c["endpoint"] = hgeEndpoint
	o, err := yaml.Marshal(c)
	Expect(err).NotTo(HaveOccurred())
	fmt.Println(o)
	// write
	err = ioutil.WriteFile(configFilePath, o, 0655)
	Expect(err).NotTo(HaveOccurred())
	return tempDir
}

func newTestNetwork(name string) (func(), error) {
	ctx := context.Background()
	reqN := testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name:           name,
			Driver:         "bridge",
			CheckDuplicate: true,
			Attachable:     true,
		},
	}

	n, err := testcontainers.GenericNetwork(ctx, reqN)
	if err != nil {
		return nil, err
	}

	removeNetwork := func() {
		err := n.Remove(ctx)
		Expect(err).NotTo(HaveOccurred())
	}
	return removeNetwork, nil
}

func randString() string {
	rand.Seed(time.Now().Unix())

	var output bytes.Buffer
	charSet := "abcdedfghijklmnopqrstABCDEFGHIJKLMNOP"
	length := 20
	for i := 0; i < length; i++ {
		random := rand.Intn(len(charSet))
		randomChar := charSet[random]
		output.WriteString(string(randomChar))
	}
	return output.String()
}
