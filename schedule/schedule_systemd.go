//+build !darwin,!windows

package schedule

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/systemd"
	"github.com/creativeprojects/resticprofile/ui"
)

func RemoveJob(profileName string) error {
	// stop the job
	cmd := exec.Command("systemctl", "--user", "stop", systemd.GetTimerFile(profileName))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	// disable the job
	cmd = exec.Command("systemctl", "--user", "disable", systemd.GetTimerFile(profileName))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	systemdUserDir, err := systemd.GetUserDir()
	if err != nil {
		return nil
	}
	timerFile := systemd.GetTimerFile(profileName)
	err = os.Remove(path.Join(systemdUserDir, timerFile))
	if err != nil {
		return nil
	}

	serviceFile := systemd.GetServiceFile(profileName)
	err = os.Remove(path.Join(systemdUserDir, serviceFile))
	if err != nil {
		return nil
	}

	return nil
}

func loadSchedules(schedules []string) ([]*calendar.Event, error) {
	events := make([]*calendar.Event, 0, len(schedules))
	for index, schedule := range schedules {
		if schedule == "" {
			return events, errors.New("empty schedule")
		}
		fmt.Printf("\nAnalyzing schedule %d/%d\n========================\n", index+1, len(schedules))
		cmd := exec.Command("systemd-analyze", "calendar", schedule)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return events, err
		}
	}
	return events, nil
}

// createJob is creating the systemd unit and activating it.
// for systemd the schedules parameter is not used.
func (j *Job) createJob() error {
	if os.Geteuid() == 0 {
		// user has sudoed already
		return j.createSystemJob()
	}
	message := "\nPlease note resticprofile was started as a standard user (typically without sudo):" +
		"\nDo you want to install the scheduled backup as a user job as opposed to a system job?"
	answer := ui.AskYesNo(os.Stdin, message, false)
	if !answer {
		return errors.New("operation cancelled")
	}
	return j.createUserJob()
}

func (j *Job) createSystemJob() error {
	return nil
}

func (j *Job) createUserJob() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	binary := absolutePathToBinary(wd, os.Args[0])
	err = systemd.Generate(wd, binary, j.configFile, j.profile.Name, j.profile.Schedule)
	if err != nil {
		return err
	}

	// enable the job
	cmd := exec.Command("systemctl", "--user", "enable", systemd.GetTimerFile(j.profile.Name))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	// start the job
	cmd = exec.Command("systemctl", "--user", "start", systemd.GetTimerFile(j.profile.Name))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (j *Job) displayStatus() error {
	cmd := exec.Command("systemctl", "--user", "status", systemd.GetTimerFile(j.profile.Name))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}