//go:build darwin

package schedule

import (
	"bytes"
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"howett.net/plist"
)

func TestPListEncoderWithCalendarInterval(t *testing.T) {
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict><key>Day</key><integer>1</integer><key>Hour</key><integer>0</integer></dict></plist>`
	entry := newCalendarInterval()
	setCalendarIntervalValueFromType(entry, 1, calendar.TypeDay)
	setCalendarIntervalValueFromType(entry, 0, calendar.TypeHour)
	buffer := &bytes.Buffer{}
	encoder := plist.NewEncoder(buffer)
	err := encoder.Encode(entry)
	require.NoError(t, err)
	assert.Equal(t, expected, buffer.String())
}

func TestGetCalendarIntervalsFromScheduleTree(t *testing.T) {
	testData := []struct {
		input    string
		expected []CalendarInterval
	}{
		{"*-*-*", []CalendarInterval{
			{"Hour": 0, "Minute": 0},
		}},
		{"*:0,30", []CalendarInterval{
			{"Minute": 0},
			{"Minute": 30},
		}},
		{"0,12:20", []CalendarInterval{
			{"Hour": 0, "Minute": 20},
			{"Hour": 12, "Minute": 20},
		}},
		{"0,12:20,40", []CalendarInterval{
			{"Hour": 0, "Minute": 20},
			{"Hour": 0, "Minute": 40},
			{"Hour": 12, "Minute": 20},
			{"Hour": 12, "Minute": 40},
		}},
		{"Mon..Fri *-*-* *:0,30:00", []CalendarInterval{
			{"Weekday": 1, "Minute": 0},
			{"Weekday": 1, "Minute": 30},
			{"Weekday": 2, "Minute": 0},
			{"Weekday": 2, "Minute": 30},
			{"Weekday": 3, "Minute": 0},
			{"Weekday": 3, "Minute": 30},
			{"Weekday": 4, "Minute": 0},
			{"Weekday": 4, "Minute": 30},
			{"Weekday": 5, "Minute": 0},
			{"Weekday": 5, "Minute": 30},
		}},
		// First sunday of the month at 3:30am
		{"Sun *-*-01..06 03:30:00", []CalendarInterval{
			{"Day": 1, "Weekday": 0, "Hour": 3, "Minute": 30},
			{"Day": 2, "Weekday": 0, "Hour": 3, "Minute": 30},
			{"Day": 3, "Weekday": 0, "Hour": 3, "Minute": 30},
			{"Day": 4, "Weekday": 0, "Hour": 3, "Minute": 30},
			{"Day": 5, "Weekday": 0, "Hour": 3, "Minute": 30},
			{"Day": 6, "Weekday": 0, "Hour": 3, "Minute": 30},
		}},
	}

	for _, testItem := range testData {
		t.Run(testItem.input, func(t *testing.T) {
			event := calendar.NewEvent()
			err := event.Parse(testItem.input)
			assert.NoError(t, err)
			assert.ElementsMatch(t, testItem.expected, getCalendarIntervalsFromScheduleTree(generateTreeOfSchedules(event)))
		})
	}
}

func TestParseStatus(t *testing.T) {
	status := `{
	"StandardOutPath" = "local.resticprofile.self.check.log";
	"LimitLoadToSessionType" = "Aqua";
	"StandardErrorPath" = "local.resticprofile.self.check.log";
	"Label" = "local.resticprofile.self.check";
	"OnDemand" = true;
	"LastExitStatus" = 0;
	"Program" = "/Users/go/src/github.com/creativeprojects/resticprofile/resticprofile";
	"ProgramArguments" = (
		"/Users/go/src/github.com/creativeprojects/resticprofile/resticprofile";
		"--no-ansi";
		"--config";
		"examples/dev.yaml";
		"--name";
		"self";
		"check";
	);
};`
	expected := map[string]string{
		"StandardOutPath":        "local.resticprofile.self.check.log",
		"LimitLoadToSessionType": "Aqua",
		"StandardErrorPath":      "local.resticprofile.self.check.log",
		"Label":                  "local.resticprofile.self.check",
		"OnDemand":               "true",
		"LastExitStatus":         "0",
		"Program":                "/Users/go/src/github.com/creativeprojects/resticprofile/resticprofile",
	}

	output := parseStatus(status)
	assert.Equal(t, expected, output)
}

func TestHandlerInstanceDefault(t *testing.T) {
	handler := NewHandler(SchedulerDefaultOS{})
	assert.NotNil(t, handler)
}

func TestHandlerInstanceLaunchd(t *testing.T) {
	handler := NewHandler(SchedulerLaunchd{})
	assert.NotNil(t, handler)
}

func TestLaunchdJobLog(t *testing.T) {
	fixtures := []struct {
		log      string
		expected string
	}{
		{"", "local.resticprofile.profile.backup.log"},
		{"udp://localhost:123", "local.resticprofile.profile.backup.log"},
		{"tcp://127.0.0.1:123", "local.resticprofile.profile.backup.log"},
		{"other file", "other file"},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.log, func(t *testing.T) {
			handler := NewHandler(SchedulerLaunchd{})
			cfg := &config.ScheduleConfig{
				Title:    "profile",
				SubTitle: "backup",
				Log:      fixture.log,
			}
			launchdJob := handler.getLaunchdJob(cfg, []*calendar.Event{})
			assert.Equal(t, fixture.expected, launchdJob.StandardOutPath)
		})
	}
}

func TestCreateUserPlist(t *testing.T) {
	handler := NewHandler(SchedulerLaunchd{})
	handler.fs = afero.NewMemMapFs()

	launchdJob := &LaunchdJob{
		Label: "TestCreateSystemPlist",
	}
	filename, err := handler.createPlistFile(launchdJob, "user")
	require.NoError(t, err)

	_, err = handler.fs.Stat(filename)
	assert.NoError(t, err)
}

func TestCreateSystemPlist(t *testing.T) {
	handler := NewHandler(SchedulerLaunchd{})
	handler.fs = afero.NewMemMapFs()

	launchdJob := &LaunchdJob{
		Label: "TestCreateSystemPlist",
	}
	filename, err := handler.createPlistFile(launchdJob, "system")
	require.NoError(t, err)

	_, err = handler.fs.Stat(filename)
	assert.NoError(t, err)
}
