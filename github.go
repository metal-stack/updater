package updater

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v56/github"
)

type release struct {
	tag      string
	date     time.Time
	url      string
	checksum string
}

func latestRelease(artefact, owner, repo string, desired *string) (*release, error) {

	client := github.NewClient(nil)

	releases, _, err := client.Repositories.ListReleases(context.Background(), owner, repo, &github.ListOptions{
		// defaults to 15, but we want to ensure we get enough releases
		// 50 is reasonably large enough to not miss any releases
		// in case we do, please refactor it yourself :)
		PerPage: 50,
	})
	if err != nil {
		return nil, err
	}

	var latestRelease *github.RepositoryRelease

	for _, r := range releases {
		if r.GetDraft() || r.GetPrerelease() {
			continue
		}
		if desired != nil && r.TagName != nil && *r.TagName != *desired {
			continue
		}
		latestRelease = r
		break
	}

	if latestRelease == nil {
		desiredVersion := "latest"
		if desired != nil {
			desiredVersion = *desired
		}
		return nil, fmt.Errorf("no release for given desired version:%q", desiredVersion)
	}

	ras, _, err := client.Repositories.ListReleaseAssets(context.Background(), owner, repo, *latestRelease.ID, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, ra := range ras {
		if ra.GetName() == artefact {
			checksum, err := slurpFile(ra.GetBrowserDownloadURL() + ".sha512")
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
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.Split(string(content), " ")[0], nil
}
