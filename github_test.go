package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLatestRelease(t *testing.T) {

	release, err := latestRelease("metalctl-linux-amd64", "metal-stack", "metalctl")
	assert.Nil(t, err)

	t.Logf("Release:%v", release)

}
