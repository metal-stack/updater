package updater

import (
	"testing"
	"time"
)

func Test_getAgeAndUptodateStatus(t *testing.T) {
	type args struct {
		latestVersionTime time.Time
		thisVersionTime   time.Time
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
				latestVersionTime: must(time.Parse(time.RFC3339, "2019-08-08T09:43:57Z")),
				thisVersionTime:   must(time.Parse(time.RFC3339, "2019-08-08T09:43:57Z")),
			},
			age:      0,
			uptodate: true,
		},
		{
			name: "same+1h",
			args: args{
				latestVersionTime: must(time.Parse(time.RFC3339, "2019-08-08T10:43:57Z")),
				thisVersionTime:   must(time.Parse(time.RFC3339, "2019-08-08T09:43:57Z")),
			},
			age:      1 * time.Hour,
			uptodate: false,
		},
		{
			name: "285h15m45s",
			args: args{
				latestVersionTime: must(time.Parse(time.RFC3339, "2019-08-20T06:59:42Z")),
				thisVersionTime:   must(time.Parse(time.RFC3339, "2019-08-08T09:43:57Z")),
			},
			age:      285*time.Hour + 15*time.Minute + 45*time.Second,
			uptodate: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAge, gotUptodate := getAgeAndUptodateStatus(tt.args.latestVersionTime, tt.args.thisVersionTime)
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
