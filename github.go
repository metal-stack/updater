package updater

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
)

type release struct {
	tag      string
	date     time.Time
	url      string
	checksum string
}

func latestRelease(artefact, owner, repo string) (*release, error) {

	client := github.NewClient(nil)

	releases, _, err := client.Repositories.ListReleases(context.Background(), owner, repo, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	var latestRelease *github.RepositoryRelease

	for _, r := range releases {
		if r.GetDraft() || r.GetPrerelease() {
			continue
		}
		latestRelease = r
		break
	}
	ras, _, err := client.Repositories.ListReleaseAssets(context.Background(), owner, repo, *latestRelease.ID, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, ra := range ras {
		if ra.GetName() == artefact {
			checksum, err := slurpFile(ra.GetBrowserDownloadURL() + ".md5")
			fmt.Printf("checksum:%s\n", checksum)
			if err != nil {
				return nil, err
			}
			return &release{
				tag:      *latestRelease.TagName,
				url:      ra.GetBrowserDownloadURL(),
				date:     latestRelease.PublishedAt.Time,
				checksum: checksum,
			}, nil
		}
	}
	return nil, fmt.Errorf("no downloadURL found for artefact:%s under %s/%s", artefact, owner, repo)
}

func slurpFile(url string) (string, error) {
	// Get the data
	//nolint:gosec,noctx
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.Split(string(content), " ")[0], nil
}
