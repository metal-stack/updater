package updater

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func Test_getAgeAndUptodateStatus(t *testing.T) {
	type args struct {
		latestVersionTag   string
		latestVersionTime  time.Time
		thisVersionVersion string
		thisVersionTime    time.Time
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
				latestVersionTime:  must(time.Parse(time.RFC3339, "2019-08-08T09:43:57Z")),
				thisVersionVersion: "v1.0.1",
				thisVersionTime:    must(time.Parse(time.RFC3339, "2019-08-08T09:43:57Z")),
			},
			age:      0,
			uptodate: true,
		},
		{
			name: "sameversion,same+1h",
			args: args{
				latestVersionTag:   "v1.0.1",
				latestVersionTime:  must(time.Parse(time.RFC3339, "2019-08-08T10:43:57Z")),
				thisVersionVersion: "v1.0.1",
				thisVersionTime:    must(time.Parse(time.RFC3339, "2019-08-08T09:43:57Z")),
			},
			age:      1 * time.Hour,
			uptodate: true,
		},
		{
			name: "minorversion,same+1h",
			args: args{
				latestVersionTag:   "v1.3.0",
				latestVersionTime:  must(time.Parse(time.RFC3339, "2019-08-08T10:43:57Z")),
				thisVersionVersion: "v1.0.1",
				thisVersionTime:    must(time.Parse(time.RFC3339, "2019-08-08T09:43:57Z")),
			},
			age:      1 * time.Hour,
			uptodate: false,
		},
		{
			name: "majorversion,same+1h",
			args: args{
				latestVersionTag:   "v2.0.0",
				latestVersionTime:  must(time.Parse(time.RFC3339, "2019-08-08T10:43:57Z")),
				thisVersionVersion: "v1.0.1",
				thisVersionTime:    must(time.Parse(time.RFC3339, "2019-08-08T09:43:57Z")),
			},
			age:      1 * time.Hour,
			uptodate: false,
		},
		{
			name: "majorversion,285h15m45s",
			args: args{
				latestVersionTag:   "v4.3.7",
				latestVersionTime:  must(time.Parse(time.RFC3339, "2019-08-20T06:59:42Z")),
				thisVersionVersion: "v3.2.1",
				thisVersionTime:    must(time.Parse(time.RFC3339, "2019-08-08T09:43:57Z")),
			},
			age:      285*time.Hour + 15*time.Minute + 45*time.Second,
			uptodate: false,
		},
		{
			name: "thisVersionNewer,same-1h",
			args: args{
				latestVersionTag:   "v1.2.3",
				latestVersionTime:  must(time.Parse(time.RFC3339, "2019-08-08T09:43:57Z")),
				thisVersionVersion: "v2.1.4",
				thisVersionTime:    must(time.Parse(time.RFC3339, "2019-08-08T10:43:57Z")),
			},
			age:      -1 * time.Hour,
			uptodate: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotAge, gotUptodate := getAgeAndUptodateStatus(tt.args.latestVersionTag, tt.args.latestVersionTime, tt.args.thisVersionVersion, tt.args.thisVersionTime)
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

	owner := "metal-stack"
	repo := "metalctl"
	programName := "metalctl"
	v := "v0.14.1"
	// Call New with appropriate arguments
	updater, err := New(owner, repo, programName, &v)

	// Check if error is nil
	if err != nil {
		t.Errorf("New returned an error: %v", err)
	}

	// Check if updater is nil
	if updater == nil {
		t.Error("New returned a nil updater")
	}

	// Check if updater fields have expected values
	expectedProgramName := "metalctl"
	if updater.programName != expectedProgramName {
		t.Errorf("Expected programName: %s, Got: %s", expectedProgramName, updater.programName)
	}
	checkSum := "084a47d1c9e7c5384855c8f93ca52852"
	if updater.checksum != checkSum {
		t.Errorf("Expected checksum: %s, Got: %s", checkSum, updater.checksum)
	}
	downUrl := "https://github.com/metal-stack/metalctl/releases/download/v0.14.1/metalctl-linux-amd64"
	if updater.downloadURL != downUrl {
		t.Errorf("Expected programName: %s, Got: %s", downUrl, updater.downloadURL)
	}
	tag := "v0.14.1"
	if updater.tag != tag {
		t.Errorf("Expected programName: %s, Got: %s", tag, updater.tag)
	}
}

func TestDownloadFunction(t *testing.T) {

	programName := "metalctl"
	tmpFile, _ := os.CreateTemp("", programName)

	url := "https://github.com/metal-stack/metalctl/releases/download/v0.14.1/metalctl-linux-amd64"
	checkSum := "084a47d1c9e7c5384855c8f93ca52852"

	downloadFile(tmpFile, url, checkSum)

	loc, _ := getOwnLocation()

	fmt.Printf("\nThis is location %v", loc)

}

type mockUpdater struct {
	downloadURL string
	checksum    string
}

func (m *mockUpdater) Do() error {
	// Mock the behavior of updating
	return nil
}

func TestUpdater_Do(t *testing.T) {

	url := "https://github.com/metal-stack/metalctl/releases/download/v0.14.1/metalctl-linux-amd64"
	checkSum := "084a47d1c9e7c5384855c8f93ca52852"
	programName := "metalctl"

	// Initialize the updater with mock data
	updater := &Updater{
		programName: programName,
		downloadURL: url,
		checksum:    checkSum,
	}

	// Run the updater
	if err := updater.Do(); err != nil {
		t.Errorf("Updater failed: %v", err)
	}

	// Add assertions as needed to verify the behavior of the updater
}

func must(tme time.Time, err error) time.Time {
	if err != nil {
		panic(err)
	}
	return tme
}
