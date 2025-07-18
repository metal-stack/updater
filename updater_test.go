package updater

import (
	"strings"
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
		artefact string
		owner    string
		repo     string
		url      string
		sums     map[string]string
	}{
		{
			name:     "metalctl@v0.18.1",
			owner:    "metal-stack",
			repo:     "metalctl",
			artefact: "metalctl",
			version:  "v0.18.1",
			sums: map[string]string{
				"linux-amd64":  "f1e483d009fc1fa95afc0f2219f13cb453a58bdcd1e86ae26e4ead220b28bb31058413a475d7ddd761c182d5e07d44232ba339dfc43d8e709169cf8e27b5c3cb",
				"darwin-arm64": "9c55822cde198fa6d37a8d05767fb2310f21179bdc91ccb28c7dd1effa0129f7a72a0e1ac995804eaee0ae7aceb1777e9ac91726d4044d9ddba80c36d45bdee5",
			},
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

			assert.Equal(t, tt.artefact, updater.programName)
			assert.Equal(t, tt.version, updater.tag)
			assert.Contains(t, updater.downloadURL, tt.version)
			assert.Contains(t, updater.downloadURL, tt.artefact)
			assert.Contains(t, updater.downloadURL, tt.owner)
			assert.Contains(t, updater.downloadURL, tt.repo)

			var found bool
			for osArch, sum := range tt.sums {
				if !strings.Contains(updater.downloadURL, osArch) {
					found = true
					continue
				}

				assert.Equal(t, sum, updater.checksum, "checksum does not match for %s", osArch)
			}

			if !found {
				t.Error("running on unknown os and arch combination", updater.downloadURL)
			}
		})
	}
}
