package main

import (
	"fmt"
	"log"
	"os"

	"github.com/go-yaml/yaml"
	"github.com/urfave/cli"

	"github.com/djmaze/shepherd/compose"
	"github.com/djmaze/shepherd/swarm"
)

// DockerAPIMinVersion is the version of the docker API, which is minimally required by
// swex. Currently we require at least API 1.26 and therefore Docker 1.13.1 or later.
const DockerAPIMinVersion string = "1.26"

var version = "master"

var (
	client swarm.Client
)

func main() {
	app := cli.NewApp()
	app.Name = "shepherd"
	app.Before = before
	app.Action = start
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "host, H",
			Usage:  "daemon socket to connect to",
			Value:  "unix:///var/run/docker.sock",
			EnvVar: "DOCKER_HOST",
		},
		cli.BoolFlag{
			Name:   "tlsverify",
			Usage:  "use TLS and verify the remote",
			EnvVar: "DOCKER_TLS_VERIFY",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func before(c *cli.Context) error {

	// configure environment vars for client
	err := envConfig(c)
	if err != nil {
		return err
	}

	return nil
}

func start(c *cli.Context) error {
	client, err := swarm.NewClient()
	if err != nil {
		panic(err)
	}

	swarmServices, err := client.ListServices()
	composeServiceCollection := compose.ServiceCollectionFromSwarmServices(swarmServices)
	stacks := composeServiceCollection.Stacks()

	for _, stack := range stacks {
		file, err := os.Create(stack.Name + ".yml")
		if err != nil {
			return err
		}

		y, err := yaml.Marshal(stack)
		if err != nil {
			return err
		}
		file.WriteString(string(y))
		file.Close()
	}

	fmt.Printf("Exported %d stacks\n", len(stacks))

	return nil
}

func setEnvOptStr(env string, opt string) error {
	if opt != "" && opt != os.Getenv(env) {
		err := os.Setenv(env, opt)
		if err != nil {
			return err
		}
	}
	return nil
}

func setEnvOptBool(env string, opt bool) error {
	if opt == true {
		return setEnvOptStr(env, "1")
	}
	return nil
}

// envConfig translates the command-line options into environment variables
// that will initialize the api client
func envConfig(c *cli.Context) error {
	var err error

	err = setEnvOptStr("DOCKER_HOST", c.GlobalString("host"))
	err = setEnvOptBool("DOCKER_TLS_VERIFY", c.GlobalBool("tlsverify"))
	err = setEnvOptStr("DOCKER_API_VERSION", DockerAPIMinVersion)

	return err
}
