package main

import (
	"fmt"
	"log"
	"strings"
	"github.com/codegangsta/cli"
)

func NewCliApp() *cli.App {

	var command string

	app := cli.NewApp()
	app.Name = "DSSH"
	app.Usage = "A distributed SSH shell"
	app.Flags = []cli.Flag {
		cli.StringFlag {
			Name:        "c, command",
			Value:       "command",
			Usage:       "the command to execute on the hosts",
			Destination: &command,
		},
	}

	app.Action = func(c *cli.Context) {

		hoststring := "localhost"

		for i := 0; i < c.NArg(); i++ {

			hoststring = c.Args()[i]
		
			parts := strings.Split(hoststring, "@")
			user := parts[0]
			host := parts[1]

			conn := NewServerConnection(user, host)
			if err := conn.Connect(); err != nil {
				log.Fatal(fmt.Errorf("Failed to connect: %s", err))
			}

			log.Println(fmt.Sprintf("%s <=================", host))
			conn.RunCommand(strings.Split(command, " "))
		}
	}

	return app
}
