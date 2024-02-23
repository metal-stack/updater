package updater

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLatestRelease(t *testing.T) {

	release, err := latestRelease("metalctl-linux-amd64", "metal-stack", "metalctl", nil)
	require.NoError(t, err)

	t.Logf("Release:%v", release)

}

func TestDesiredRelease(t *testing.T) {

	v := "v0.14.1"
	check := "084a47d1c9e7c5384855c8f93ca52852"
	fullNameCheck := "metalctl" + "-" + runtime.GOOS + "-" + runtime.GOARCH
	release, err := latestRelease(fullNameCheck, "metal-stack", "metalctl", &v)
	require.NoError(t, err)

	t.Logf("\nRelease:%v", release)
	t.Logf("\nReleaseURL:%v", release.url)
	t.Logf("\nReleaseChecksum:%v", release.checksum)
	t.Logf("\nReleaseDate:%v", release.date)
	t.Logf("\nReleaseTag:%v", release.tag)
	require.Equal(t, v, release.tag)
	require.Equal(t, check, release.checksum)
	require.Contains(t, release.url, "v0.14.1")
}
