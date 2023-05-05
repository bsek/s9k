package github

import (
	"context"
	"fmt"
	"time"

	"github.com/bsek/s9k/internal/s9k/utils"
	"github.com/google/go-github/v50/github"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

var client *github.Client

type Commit struct {
	Sha     string
	Message string
}

type Package struct {
	Sha     string
	Image   string
	Created time.Time
}

func CreateClient(token string) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: utils.RemoveAllRegex("\n", token)},
	)
	tc := oauth2.NewClient(ctx, ts)

	client = github.NewClient(tc)
}

func FetchPackagesfromGhcr(name string) ([]Package, error) {
	opts := github.PackageListOptions{
		PackageType: github.String("container"),
		ListOptions: github.ListOptions{
			PerPage: 10,
		},
	}

	versions, _, err := client.Organizations.PackageGetAllVersions(context.Background(), "oslokommune", "container", name, &opts)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to read packages from repo %s", name)
		return nil, err
	}

	var list = make([]Package, 0, len(versions))
	for _, v := range versions {
		var tag string
		if len(v.Metadata.Container.Tags) > 0 {
			tag = v.Metadata.Container.Tags[0]

			shortSha := tag[len(tag)-7:]

			list = append(list, Package{
				Sha:     shortSha,
				Image:   fmt.Sprintf("ghcr.io/oslokommune/%s:%s", name, tag),
				Created: v.CreatedAt.Time,
			})
		}
	}

	return list, nil
}

func FetchCommits(name string) ([]Commit, error) {
	opts := github.CommitsListOptions{
		SHA:         "master",
		ListOptions: github.ListOptions{PerPage: 20},
	}
	commits, _, err := client.Repositories.ListCommits(context.Background(), "oslokommune", name, &opts)

	if err != nil {
		log.Error().Err(err).Msgf("Failed to read commits from repo %s", name)
		return nil, err
	}

	var list = make([]Commit, 0, len(commits))
	for _, v := range commits {
		msg := ""
		commit := v.Commit
		if commit != nil {
			msg = *commit.Message
		}

		list = append(list, Commit{
			Sha:     *v.SHA,
			Message: msg,
		})
	}

	return list, nil
}

func CallGithubAction(clusterName, serviceName, version string) error {
	owner := "oslokommune"
	repo := "skjema-iac-terraform"
	workflowName := "update_container_image_version.yml"

	payload := map[string]interface{}{
		"application":   serviceName,
		"image_version": fmt.Sprintf("ghcr.io/oslokommune/%s", version),
		"environment":   clusterName,
	}

	event := github.CreateWorkflowDispatchEventRequest{
		Ref:    "master",
		Inputs: payload,
	}

	_, err := client.Actions.CreateWorkflowDispatchEventByFileName(context.Background(), owner, repo, workflowName, event)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to invoke github action %s", workflowName)
		return err
	}

	return nil
}
