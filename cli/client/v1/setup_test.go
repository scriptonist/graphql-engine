package v1

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"
)

var hasura = flag.Bool("hasura", false, "run hasura integration tests")

func TestMain(m *testing.M) {
	flag.Parse()
	var teardown func() error
	var err error
	if *hasura {
		// setup a hasura instance
		if teardown, err = setupHasuraDockerCompose(); err != nil {
			panic(err)
		}
	}
	// run tests
	result := m.Run()

	if *hasura {
		// teardown the hasura instance
		teardown()
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
		log.Println("tearing down test assets")
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
