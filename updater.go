package updater

import (
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/metal-stack/v"

	"github.com/cheggaaa/pb/v3"

	"github.com/blang/semver/v4"
)

// Updater update a running binary
type Updater struct {
	programName string
	downloadURL string
	checksum    string
	date        time.Time
	tag         string
}

// New create a Updater
func New(owner, repo, programName string) (*Updater, error) {

	fullProgramName := programName + "-" + runtime.GOOS + "-" + runtime.GOARCH

	release, err := latestRelease(fullProgramName, owner, repo)
	if err != nil {
		return nil, err
	}

	return &Updater{
		programName: programName,
		downloadURL: release.url,
		checksum:    release.checksum,
		date:        release.date,
		tag:         release.tag,
	}, nil
}

// Do actually updates local programm with the most recent found on the download server
func (u *Updater) Do() error {

	tmpFile, err := ioutil.TempFile("", u.programName)
	if err != nil {
		return fmt.Errorf("unable create tempfile:%w", err)
	}
	err = downloadFile(tmpFile, u.downloadURL, u.checksum)
	if err != nil {
		return err
	}
	location, err := getOwnLocation()
	if err != nil {
		return fmt.Errorf("unable to get own binary location:%w", err)
	}
	info, err := os.Stat(location)
	if err != nil {
		return fmt.Errorf("unable to stat old binary:%w", err)
	}
	mode := info.Mode()
	lf, err := os.OpenFile(location, os.O_WRONLY, mode)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("unable to write to:%s need root access:%w", location, err)
		}
	}
	lf.Close()

	oldlocation := location + ".update"
	defer os.Remove(oldlocation)
	err = os.Rename(location, oldlocation)
	if err != nil {
		return fmt.Errorf("unable to rename old binary:%w", err)
	}

	err = copy(tmpFile.Name(), location)
	if err != nil {
		return fmt.Errorf("unable to copy:%w", err)
	}
	err = os.Chmod(location, mode)
	if err != nil {
		return fmt.Errorf("unable to chown:%w", err)
	}

	return nil
}

// Check version of locally installed program with that available on the download server.
func (u *Updater) Check() error {
	thisVersionBuildtime, err := time.Parse(time.RFC3339, v.BuildDate)
	if err != nil {
		return err
	}

	location, err := getOwnLocation()
	if err != nil {
		return err
	}

	fmt.Printf("latest version:%s from:%s\n", u.tag, u.date.Format(time.RFC3339))
	fmt.Printf("local  version:%s from:%s\n", v.Version, thisVersionBuildtime.Format(time.RFC3339))

	age, isUpToDate := getAgeAndUptodateStatus(u.tag, u.date, v.Version, thisVersionBuildtime)
	if isUpToDate {
		fmt.Printf("%s is up to date\n", u.programName)
	} else {
		fmt.Printf("%s is %s old, please run '%s update do'\n", u.programName, humanizeDuration(age), u.programName)
		fmt.Printf("%s location:%s\n", u.programName, location)
	}

	return nil
}

// getAgeAndStatus calculates the age (difference in release time) and decides if the local version is up to date based on semantic version comparison, returns true if up to date.
func getAgeAndUptodateStatus(latestVersionTag string, latestVersionTime time.Time, thisVersionVersion string, thisVersionTime time.Time) (time.Duration, bool) {

	// latestVersionBuildtime is expected to be "greater" or "equal" than the version of this binary "thisVersionBuildtime"
	age := latestVersionTime.Sub(thisVersionTime)

	latestVersion, err := semver.ParseTolerant(latestVersionTag)
	if err != nil {
		fmt.Printf("Error: Latest version tag %s is not a valid semantic version!\n", latestVersionTag)
		return 0, false
	}

	thisVersion, err := semver.ParseTolerant(thisVersionVersion)
	if err != nil {
		fmt.Printf("Error: Local version string %s is not a valid semantic version!\n", thisVersionVersion)
		return 0, false
	}

	return age, semver.Version.LE(latestVersion, thisVersion)
}

func getOwnLocation() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	location, err := filepath.EvalSymlinks(ex)
	if err != nil {
		return "", err
	}
	return location, nil
}

func md5sum(binary string) (string, error) {
	//nolint:gosec
	hasher := md5.New()
	s, err := ioutil.ReadFile(binary)
	if err != nil {
		return "", err
	}
	_, err = hasher.Write(s)
	if err != nil {
		return "", err
	}

	return string(hex.EncodeToString(hasher.Sum(nil))), nil
}

// downloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func downloadFile(out *os.File, url, checksum string) error {

	// Get the data
	//nolint:gosec,noctx
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	defer out.Close()
	fileSize := resp.ContentLength

	bar := pb.Full.Start64(fileSize)
	// create proxy reader
	barReader := bar.NewProxyReader(resp.Body)
	_, err = io.Copy(out, barReader)
	if err != nil {
		return err
	}
	bar.Finish()

	c, err := md5sum(out.Name())
	if err != nil {
		return fmt.Errorf("unable to calculate checksum:%w", err)
	}
	if c != checksum {
		return fmt.Errorf("checksum mismatch %s:%s", c, checksum)
	}

	return err
}

func copy(src, dst string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dst, input, os.ModeType)
	if err != nil {
		return err
	}
	return nil
}

func humanizeDuration(duration time.Duration) string {
	days := int64(duration.Hours() / 24)
	hours := int64(math.Mod(duration.Hours(), 24))
	minutes := int64(math.Mod(duration.Minutes(), 60))
	seconds := int64(math.Mod(duration.Seconds(), 60))

	chunks := []struct {
		singularName string
		amount       int64
	}{
		{"d", days},
		{"h", hours},
		{"m", minutes},
		{"s", seconds},
	}

	parts := []string{}

	for _, chunk := range chunks {
		switch chunk.amount {
		case 0:
			continue
		default:
			parts = append(parts, fmt.Sprintf("%d%s", chunk.amount, chunk.singularName))
		}
	}

	if len(parts) == 0 {
		return "0s"
	}
	if len(parts) > 2 {
		parts = parts[:2]
	}
	return strings.Join(parts, " ")
}
