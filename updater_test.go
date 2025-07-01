package updater

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getAgeAndUptodateStatus(t *testing.T) {
	type args struct {
		latestVersionTag   string
		latestVersionTime  string
		thisVersionVersion string
		thisVersionTime    string
	}
	tests := []struct {
		name     string
		args     args
		age      time.Duration
		uptodate bool
	}{
		{
			name: "same",
			args: args{
				latestVersionTag:   "v1.0.1",
				latestVersionTime:  "2019-08-08T09:43:57Z",
				thisVersionVersion: "v1.0.1",
				thisVersionTime:    "2019-08-08T09:43:57Z",
			},
			age:      0,
			uptodate: true,
		},
		{
			name: "sameversion,same+1h",
			args: args{
				latestVersionTag:   "v1.0.1",
				latestVersionTime:  "2019-08-08T10:43:57Z",
				thisVersionVersion: "v1.0.1",
				thisVersionTime:    "2019-08-08T09:43:57Z",
			},
			age:      1 * time.Hour,
			uptodate: true,
		},
		{
			name: "minorversion,same+1h",
			args: args{
				latestVersionTag:   "v1.3.0",
				latestVersionTime:  "2019-08-08T10:43:57Z",
				thisVersionVersion: "v1.0.1",
				thisVersionTime:    "2019-08-08T09:43:57Z",
			},
			age:      1 * time.Hour,
			uptodate: false,
		},
		{
			name: "majorversion,same+1h",
			args: args{
				latestVersionTag:   "v2.0.0",
				latestVersionTime:  "2019-08-08T10:43:57Z",
				thisVersionVersion: "v1.0.1",
				thisVersionTime:    "2019-08-08T09:43:57Z",
			},
			age:      1 * time.Hour,
			uptodate: false,
		},
		{
			name: "majorversion,285h15m45s",
			args: args{
				latestVersionTag:   "v4.3.7",
				latestVersionTime:  "2019-08-20T06:59:42Z",
				thisVersionVersion: "v3.2.1",
				thisVersionTime:    "2019-08-08T09:43:57Z",
			},
			age:      285*time.Hour + 15*time.Minute + 45*time.Second,
			uptodate: false,
		},
		{
			name: "thisVersionNewer,same-1h",
			args: args{
				latestVersionTag:   "v1.2.3",
				latestVersionTime:  "2019-08-08T09:43:57Z",
				thisVersionVersion: "v2.1.4",
				thisVersionTime:    "2019-08-08T10:43:57Z",
			},
			age:      -1 * time.Hour,
			uptodate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			latestVersionTime, err := time.Parse(time.RFC3339, tt.args.latestVersionTime)
			require.NoError(t, err)

			thisVersionTime, err := time.Parse(time.RFC3339, tt.args.thisVersionTime)
			require.NoError(t, err)

			gotAge, gotUptodate := getAgeAndUptodateStatus(tt.args.latestVersionTag, latestVersionTime, tt.args.thisVersionVersion, thisVersionTime)
			if gotAge != tt.age {
				t.Errorf("getAgeAndUptodateStatus() gotAge = %v, want %v", gotAge, tt.age)
			}
			if gotUptodate != tt.uptodate {
				t.Errorf("getAgeAndUptodateStatus() gotUptodate = %v, want %v", gotUptodate, tt.uptodate)
			}
		})
	}
}

func TestNewUpdater(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		sum      string
		artefact string
		owner    string
		repo     string
		url      string
		fullName string // optional override
	}{
		{
			name:     "metalctl@v0.18.0 linux",
			owner:    "metal-stack",
			repo:     "metalctl",
			artefact: "metalctl",
			fullName: "metalctl-linux-amd64",
			version:  "v0.18.0",
			sum:      "f423e891ba1034242913b030cc0692c1",
			url:      "https://github.com/metal-stack/metalctl/releases/download/v0.18.0/metalctl-linux-amd64",
		},
		{
			name:     "metalctl@v0.18.0 arm mac",
			owner:    "metal-stack",
			repo:     "metalctl",
			artefact: "metalctl",
			fullName: "metalctl-darwin-arm64",
			version:  "v0.18.0",
			sum:      "662284f6f9f015bd0b55c36ec3aae26e",
			url:      "https://github.com/metal-stack/metalctl/releases/download/v0.18.0/metalctl-darwin-arm64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater, err := New(
				tt.owner,
				tt.repo,
				tt.artefact,
				&tt.version,
			)
			require.NoError(t, err)

			if tt.fullName != "" {
				updater.programName = tt.fullName
			}

			assert.Equal(t, tt.artefact, updater.programName)
			assert.Equal(t, tt.sum, updater.checksum)
			assert.Equal(t, tt.version, updater.tag)
			assert.Equal(t, tt.url, updater.downloadURL)
		})
	}
}
