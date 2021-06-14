package updater

import (
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

func must(tme time.Time, err error) time.Time {
	if err != nil {
		panic(err)
	}
	return tme
}
