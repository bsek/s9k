package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/bsek/s9k/internal/s9k/application"
	"github.com/bsek/s9k/internal/s9k/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	file, err := os.OpenFile(
		"myapp.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	if err != nil {
		panic(err)
	}
	log.Logger = zerolog.New(file).With().Timestamp().Logger()

	builder := new(strings.Builder)

	// retrieve github token from gh tool
	c := exec.Command("gh", "auth", "token")
	c.Stdout = builder
	c.Stderr = builder
	err = c.Run()

	if err != nil {
		panic(err.Error())
	}

	// create github client
	github.CreateClient(builder.String())

	flag.Usage = func() {
		fmt.Printf("Usage: %s\n\n%s uses your valid AWS session credentials to display a visual inspection of your account's ECS clusters and Lambda functions.\n", "sk", "sk")
	}
	flag.Parse()

	application.Entrypoint()
}
