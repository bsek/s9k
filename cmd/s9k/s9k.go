package main

import (
	"flag"
	"os"
	"os/exec"
	"strings"

	"github.com/bsek/s9k/internal/entrypoint"
	"github.com/bsek/s9k/internal/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	debug := flag.Bool("debug", false, "sets log level to debug")

	flag.Parse()

	// Default level is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	file, err := os.OpenFile(
		"/tmp/s9k.log",
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
		log.Fatal().Err(err).Msg("Failed to load token from gh tool")
	}

	// create github client
	github.CreateClient(builder.String())

	entrypoint.Entrypoint()
}
