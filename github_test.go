package updater

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLatestRelease(t *testing.T) {

	release, err := latestRelease("metalctl-linux-amd64", "metal-stack", "metalctl")
	require.NoError(t, err)

	t.Logf("Release:%v", release)

}
