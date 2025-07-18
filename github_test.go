package updater

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLatestRelease(t *testing.T) {
	release, err := latestRelease("metalctl-linux-amd64", "metal-stack", "metalctl", nil)
	require.NoError(t, err)

	t.Logf("Release:%v", release)
}

func TestDesiredRelease(t *testing.T) {
	v := "v0.18.0"
	release, err := latestRelease("metalctl-linux-amd64", "metal-stack", "metalctl", &v)
	require.NoError(t, err)

	t.Logf("Release:%v", release)
	require.Equal(t, v, release.tag)
	require.Contains(t, release.url, "v0.18.0")
}

func TestUnknownDesiredRelease(t *testing.T) {
	v := "v0.0.1"
	release, err := latestRelease("metalctl-linux-amd64", "metal-stack", "metalctl", &v)
	require.Error(t, err, "no release for given desired version:\"v0.0.1\"")
	require.Nil(t, release)
}
