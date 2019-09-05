package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

	"github.com/metal-pod/v"
	"github.com/pkg/errors"

	"github.com/cheggaaa/pb/v3"
)

// Updater update a running binary
type Updater struct {
	programName       string
	downloadURLPrefix string
	binaryURL         string
	releaseURL        string
}

// New create a Updater
func New(downloadURLPrefix, programName string) *Updater {
	return &Updater{
		programName:       programName,
		downloadURLPrefix: downloadURLPrefix,
		releaseURL:        downloadURLPrefix + "version-" + runtime.GOOS + "-" + runtime.GOARCH + ".json",
		binaryURL:         downloadURLPrefix + programName + "-" + runtime.GOOS + "-" + runtime.GOARCH,
	}
}

// Release represents a release
type Release struct {
	Version  time.Time
	Checksum string
}

// Do actually updates local programm with the most recent found on the download server
func (u *Updater) Do() error {
	latestVersion, err := u.versionInfo(u.releaseURL)
	if err != nil {
		return fmt.Errorf("unable read version information:%v", err)
	}

	tmpFile, err := ioutil.TempFile("", u.programName)
	if err != nil {
		return fmt.Errorf("unable create tempfile:%v", err)
	}
	err = downloadFile(tmpFile, u.binaryURL, latestVersion.Checksum)
	if err != nil {
		return err
	}
	location, err := getOwnLocation()
	if err != nil {
		return fmt.Errorf("unable to get own binary location:%v", err)
	}
	info, err := os.Stat(location)
	if err != nil {
		return fmt.Errorf("unable to stat old binary:%v", err)
	}
	mode := info.Mode()
	lf, err := os.OpenFile(location, os.O_WRONLY, mode)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("unable to write to:%s need root access:%v", location, err)
		}
	}
	lf.Close()

	oldlocation := location + ".update"
	defer os.Remove(oldlocation)
	err = os.Rename(location, oldlocation)
	if err != nil {
		return fmt.Errorf("unable to rename old binary:%v", err)
	}

	err = copy(tmpFile.Name(), location)
	if err != nil {
		return fmt.Errorf("unable to copy:%v", err)
	}
	err = os.Chmod(location, mode)
	if err != nil {
		return fmt.Errorf("unable to chown:%v", err)
	}

	return nil
}

// Dump version of local binary in json format which is suitable as version.json locaten on the downloadserver.
func (u *Updater) Dump(fullPath string) error {
	if len(fullPath) < 1 {
		return errors.New("program binary as first argument required")
	}

	checksum, err := sum(fullPath)
	if err != nil {
		return err
	}
	version, err := time.Parse(time.RFC3339, v.BuildDate)
	if err != nil {
		return err
	}
	r := Release{
		Version:  version,
		Checksum: checksum,
	}
	bytes, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Println(string(bytes))
	return err
}

// Check version of locally installed program with that available on the download server.
func (u *Updater) Check() error {
	latestVersionBuildtime, err := u.versionInfo(u.releaseURL)
	if err != nil {
		return err
	}
	thisVersionBuildtime, err := time.Parse(time.RFC3339, v.BuildDate)
	if err != nil {
		return err
	}

	location, err := getOwnLocation()
	if err != nil {
		return err
	}

	fmt.Printf("latest version:%s\n", latestVersionBuildtime.Version.Format(time.RFC3339))
	fmt.Printf("local  version:%s\n", thisVersionBuildtime.Format(time.RFC3339))

	age, isUpToDate := getAgeAndUptodateStatus(latestVersionBuildtime.Version, thisVersionBuildtime)
	if isUpToDate {
		fmt.Printf("%s is up to date\n", u.programName)
	} else {
		fmt.Printf("%s is %s old, please run '%s update do'\n", u.programName, humanizeDuration(age), u.programName)
		fmt.Printf("%s location:%s\n", u.programName, location)
	}

	return nil
}

// getAgeAndStatus calculates the age and decides if it is considered up to date, returns true if uptodate
func getAgeAndUptodateStatus(latestVersionTime, thisVersionTime time.Time) (time.Duration, bool) {

	// latestVersionBuildtime is expected to be "greater" or "equal" than the version of this binary "thisVersionBuildtime"
	age := latestVersionTime.Sub(thisVersionTime)

	return age, age <= 0
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

func sum(binary string) (string, error) {
	hasher := sha256.New()
	s, err := ioutil.ReadFile(binary)
	if err != nil {
		return "", err
	}
	hasher.Write(s)
	if err != nil {
		return "", err
	}

	return string(hex.EncodeToString(hasher.Sum(nil))), nil
}

func (u *Updater) versionInfo(url string) (Release, error) {
	var release Release
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return release, errors.Wrap(err, "error creating new http request")
	}

	resp, err := client.Do(req)
	if err != nil {
		return release, errors.Wrapf(err, "error with http GET for endpoint %s", url)
	}

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return release, errors.Wrap(err, "Error getting json from "+u.programName+" version url")
	}
	return release, nil
}

// downloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func downloadFile(out *os.File, url, checksum string) error {

	// Get the data
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
	bar.Finish()

	c, err := sum(out.Name())
	if err != nil {
		return fmt.Errorf("unable to calculate checksum:%v", err)
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
