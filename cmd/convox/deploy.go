package main

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/convox/rack/cmd/convox/stdcli"
)

func init() {
	stdcli.RegisterCommand(cli.Command{
		Name:        "deploy",
		Description: "deploy an app to AWS",
		Usage:       "<directory>",
		Action:      cmdDeploy,
		Flags: []cli.Flag{
			appFlag,
			cli.BoolFlag{
				Name:  "no-cache",
				Usage: "Do not use Docker cache during build.",
			},
		},
	})
}

func cmdDeploy(c *cli.Context) {
	wd := "."

	if len(c.Args()) > 0 {
		wd = c.Args()[0]
	}

	dir, app, err := stdcli.DirApp(c, wd)

	if err != nil {
		stdcli.Error(err)
		return
	}

	fmt.Printf("Deploying %s\n", app)

	a, err := rackClient(c).GetApp(app)

	if err != nil {
		stdcli.Error(err)
		return
	}

	switch a.Status {
	case "creating":
		stdcli.Error(fmt.Errorf("app is still creating: %s", app))
		return
	case "running", "updating":
	default:
		stdcli.Error(fmt.Errorf("unable to build app: %s", app))
		return
	}

	// build
	release, err := executeBuild(c, dir, app)

	if err != nil {
		stdcli.Error(err)
		return
	}

	if release == "" {
		return
	}

	fmt.Printf("Promoting %s... ", release)

	_, err = rackClient(c).PromoteRelease(app, release)

	if err != nil {
		stdcli.Error(err)
		return
	}

	fmt.Println("UPDATING")

	formation, err := rackClient(c).ListFormation(app)

	if err != nil {
		stdcli.Error(err)
		return
	}

	for _, processType := range formation {
		for _, port := range processType.Ports {
			fmt.Printf("%s will be available at %s:%d\n", processType.Name, a.Balancer, port)
		}
	}
}
