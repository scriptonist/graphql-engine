package seed

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

var hasura = flag.Bool("hasura", false, "run hasura integration tests")
var reuse = flag.Bool("reuse", false, "reuse test assets")
var noteardown = flag.Bool("no-teardown", false, "don't teardown test assets after test")

func TestMain(m *testing.M) {
	flag.Parse()
	var teardown func() error
	var err error
	if *hasura && !*reuse {
		log.Println("setting up test assets")
		// setup a hasura instance
		if teardown, err = setupHasuraDockerCompose(); err != nil {
			panic(err)
		}

		var retries = 20
		// wait for hasura to get ready
		log.Printf("waiting for hasura")
		for retries > 0 {
			_, err := http.Get("http://0.0.0.0:8080/healthz")
			if err != nil {
				retries--
				time.Sleep(time.Second * 5)
				continue
			}
			break
		}
		log.Println("Hasura up")

	}
	// run tests
	result := m.Run()

	// teardown test assets when
	// noteardown and reuse flag is not set
	if *hasura && !*noteardown && !*reuse {
		log.Println("tearing down test assets")
		// teardown the hasura instance
		if err := teardown(); err != nil {
			log.Printf("teardown failed: %v", err)
		}
	}
	os.Exit(result)
}

func setupHasuraDockerCompose() (func() error, error) {
	const dataDirectory = "./testdata/hasuradata"
	// Get dockerfile
	commands := fmt.Sprintf("rm -rf %s; mkdir -p %s ; cd %s ; curl -LO https://raw.githubusercontent.com/hasura/graphql-engine/master/install-manifests/docker-compose/docker-compose.yaml", dataDirectory, dataDirectory, dataDirectory)
	cmd := exec.Command("bash", "-c", commands)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err, string(out))
		return nil, err
	}

	commands = `docker-compose up -d`
	cmd = exec.Command("bash", "-c", commands)
	cmd.Dir = dataDirectory
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return nil, err
	}

	teardown := func() error {
		commands = `docker-compose down -v`
		cmd = exec.Command("bash", "-c", commands)
		cmd.Dir = dataDirectory
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Println(string(out))
			return err
		}

		cmd = exec.Command("bash", "-c", "rm -rf ./testdata")
		out, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Println(out)
			return err
		}
		return nil
	}

	return teardown, nil
}
