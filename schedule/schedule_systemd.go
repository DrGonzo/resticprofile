//+build !darwin,!windows

package schedule

import (
	"errors"
	"os"
	"os/exec"
	"path"

	"github.com/creativeprojects/resticprofile/systemd"
	"github.com/creativeprojects/resticprofile/ui"
)

const (
	systemdBin   = "systemd"
	systemctlBin = "systemctl"
)

// checkSystem verifies systemd is available on this system
func checkSystem() error {
	found, err := exec.LookPath(systemdBin)
	if err != nil || found == "" {
		return errors.New("it doesn't look like systemd is installed on your system")
	}
	return nil
}

// removeJob is disabling the systemd unit and deleting the timer and service files
func (j *Job) removeJob() error {
	if os.Geteuid() == 0 {
		// user has sudoed
		return j.removeSystemJob()
	}
	return j.removeUserJob()
}

// removeJob is disabling the systemd unit and deleting the timer and service files
func (j *Job) removeSystemJob() error {
	// stop the job
	cmd := exec.Command(systemctlBin, "stop", systemd.GetTimerFile(j.profile.Name))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	// disable the job
	cmd = exec.Command(systemctlBin, "disable", systemd.GetTimerFile(j.profile.Name))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	timerFile := systemd.GetTimerFile(j.profile.Name)
	err = os.Remove(path.Join(systemd.GetSystemDir(), timerFile))
	if err != nil {
		return nil
	}

	serviceFile := systemd.GetServiceFile(j.profile.Name)
	err = os.Remove(path.Join(systemd.GetSystemDir(), serviceFile))
	if err != nil {
		return nil
	}

	return nil
}

// removeJob is disabling the systemd unit and deleting the timer and service files
func (j *Job) removeUserJob() error {
	// stop the job
	cmd := exec.Command(systemctlBin, "--user", "stop", systemd.GetTimerFile(j.profile.Name))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	// disable the job
	cmd = exec.Command(systemctlBin, "--user", "disable", systemd.GetTimerFile(j.profile.Name))
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
	timerFile := systemd.GetTimerFile(j.profile.Name)
	err = os.Remove(path.Join(systemdUserDir, timerFile))
	if err != nil {
		return nil
	}

	serviceFile := systemd.GetServiceFile(j.profile.Name)
	err = os.Remove(path.Join(systemdUserDir, serviceFile))
	if err != nil {
		return nil
	}

	return nil
}

// createJob is creating the systemd unit and activating it
func (j *Job) createJob() error {
	if os.Geteuid() == 0 {
		// user has sudoed already
		return j.createSystemJob()
	}
	message := "\nPlease note resticprofile was started as a standard user (typically without sudo):" +
		"\nDo you want to install the scheduled backup as a user job as opposed to a system job?"
	answer := ui.AskYesNo(os.Stdin, message, false)
	if !answer {
		return errors.New("operation cancelled by user")
	}
	return j.createUserJob()
}

func (j *Job) createSystemJob() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	binary, err := os.Executable()
	if err != nil {
		return err
	}

	err = systemd.Generate(wd, binary, j.configFile, j.profile.Name, j.profile.Schedule, systemd.SystemUnit)
	if err != nil {
		return err
	}

	// enable the job
	cmd := exec.Command(systemctlBin, "enable", systemd.GetTimerFile(j.profile.Name))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	// start the job
	cmd = exec.Command(systemctlBin, "start", systemd.GetTimerFile(j.profile.Name))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (j *Job) createUserJob() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	binary, err := os.Executable()
	if err != nil {
		return err
	}

	err = systemd.Generate(wd, binary, j.configFile, j.profile.Name, j.profile.Schedule, systemd.UserUnit)
	if err != nil {
		return err
	}

	// enable the job
	cmd := exec.Command(systemctlBin, "--user", "enable", systemd.GetTimerFile(j.profile.Name))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	// start the job
	cmd = exec.Command(systemctlBin, "--user", "start", systemd.GetTimerFile(j.profile.Name))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (j *Job) displayStatus() error {
	cmd := exec.Command(systemctlBin, "--user", "status", systemd.GetTimerFile(j.profile.Name))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
