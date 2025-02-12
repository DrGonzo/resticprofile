package filesearch

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Quick test to see the error message on the travis build agents:
//
// Linux:
// filesearch_test.go:13: could not locate `some_file` in any of the following paths: /home/travis/.config/some_service/some_path, /etc/xdg/some_service/some_path
//
// macOS:
// filesearch_test.go:13: could not locate `some_file` in any of the following paths: /Users/travis/Library/Preferences/some_service/some_path, /Library/Preferences/some_service/some_path
//
// Windows:
// filesearch_test.go:13: could not locate `some_file` in any of the following paths: C:\Users\travis\AppData\Local\some_service\some_path, C:\ProgramData\some_service\some_path
func TestSearchConfigFile(t *testing.T) {
	found, err := xdg.SearchConfigFile(filepath.Join("some_service", "some_path", "some_file"))
	t.Log(err)
	assert.Empty(t, found)
	assert.Error(t, err)
}

// Quick test to see the default xdg config on the build agents
//
// Linux:
// ConfigHome: /home/travis/.config
// ConfigDirs: [/etc/xdg]
//
// macOS:
// ConfigHome: /Users/travis/Library/Preferences
// ConfigDirs: [/Library/Preferences]
//
// Windows:
// ConfigHome: C:\Users\travis\AppData\Local
// ConfigDirs: [C:\ProgramData]
func TestDefaultConfigDirs(t *testing.T) {
	t.Log("ConfigHome:", xdg.ConfigHome)
	t.Log("ConfigDirs:", xdg.ConfigDirs)
}

type testLocation struct {
	realPath        string
	realFile        string
	searchPath      string
	searchFile      string
	deletePathAfter bool
}

func TestFindConfigurationFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skip this test in short mode")
	}
	// Work from a temporary directory
	err := os.Chdir(os.TempDir())
	require.NoError(t, err)
	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Log("Working directory:", cwd)

	locations := []testLocation{
		{
			realPath:   "",
			realFile:   "profiles.spec",
			searchPath: "",
			searchFile: "profiles.spec",
		},
		{
			realPath:   "",
			realFile:   "profiles.conf",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:   "",
			realFile:   "profiles.yaml",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:   "",
			realFile:   "profiles.json",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:   "",
			realFile:   "profiles.toml",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:   "",
			realFile:   "profiles.hcl",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:        "unittest-config",
			realFile:        "profiles.spec",
			searchPath:      "unittest-config",
			searchFile:      "profiles.spec",
			deletePathAfter: true,
		},
		{
			realPath:        "unittest-config",
			realFile:        "profiles.conf",
			searchPath:      "unittest-config",
			searchFile:      "profiles",
			deletePathAfter: true,
		},
		{
			realPath:        "unittest-config",
			realFile:        "profiles.toml",
			searchPath:      "unittest-config",
			searchFile:      "profiles",
			deletePathAfter: true,
		},
		{
			realPath:        "unittest-config",
			realFile:        "profiles.yaml",
			searchPath:      "unittest-config",
			searchFile:      "profiles",
			deletePathAfter: true,
		},
		{
			realPath:        "unittest-config",
			realFile:        "profiles.json",
			searchPath:      "unittest-config",
			searchFile:      "profiles",
			deletePathAfter: true,
		},
		{
			realPath:        "unittest-config",
			realFile:        "profiles.hcl",
			searchPath:      "unittest-config",
			searchFile:      "profiles",
			deletePathAfter: true,
		},
		{
			realPath:   filepath.Join(xdg.ConfigHome, "resticprofile"),
			realFile:   "profiles.spec",
			searchPath: "",
			searchFile: "profiles.spec",
		},
		{
			realPath:   filepath.Join(xdg.ConfigHome, "resticprofile"),
			realFile:   "profiles.conf",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:   filepath.Join(xdg.ConfigHome, "resticprofile"),
			realFile:   "profiles.toml",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:   filepath.Join(xdg.ConfigHome, "resticprofile"),
			realFile:   "profiles.yaml",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:   filepath.Join(xdg.ConfigHome, "resticprofile"),
			realFile:   "profiles.json",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:   filepath.Join(xdg.ConfigHome, "resticprofile"),
			realFile:   "profiles.hcl",
			searchPath: "",
			searchFile: "profiles",
		},
	}
	for _, location := range locations {
		var err error
		// Install empty config file
		if location.realPath != "" {
			err = os.MkdirAll(location.realPath, 0700)
			require.NoError(t, err)
		}
		file, err := os.Create(filepath.Join(location.realPath, location.realFile))
		require.NoError(t, err)
		file.Close()

		// Test
		found, err := FindConfigurationFile(filepath.Join(location.searchPath, location.searchFile))
		assert.NoError(t, err)
		assert.NotEmpty(t, found)
		assert.Equal(t, filepath.Join(location.realPath, location.realFile), found)

		// Clears up the test file
		if location.realPath == "" || !location.deletePathAfter {
			err = os.Remove(filepath.Join(location.realPath, location.realFile))
		} else {
			err = os.RemoveAll(location.realPath)
		}
		require.NoError(t, err)
	}
}

func TestCannotFindConfigurationFile(t *testing.T) {
	found, err := FindConfigurationFile("some_config_file")
	assert.Empty(t, found)
	assert.Error(t, err)
}

func TestFindResticBinary(t *testing.T) {
	binary, err := FindResticBinary("some_other_name")
	if binary != "" {
		assert.True(t, strings.HasSuffix(binary, getResticBinary()))
		assert.NoError(t, err)
	} else {
		assert.Error(t, err)
	}
}

func TestFindResticBinaryWithTilde(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on Windows")
		return
	}
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tempFile, err := os.CreateTemp(home, "TestFindResticBinaryWithTilde")
	require.NoError(t, err)
	tempFile.Close()
	defer func() {
		os.Remove(tempFile.Name())
	}()

	search := filepath.Join("~", filepath.Base(tempFile.Name()))
	binary, err := FindResticBinary(search)
	require.NoError(t, err)
	assert.Equalf(t, tempFile.Name(), binary, "cannot find %q", search)
}

func TestShellExpand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on Windows")
		return
	}
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	user, err := user.Current()
	require.NoError(t, err)

	testData := []struct {
		source   string
		expected string
	}{
		{"/", "/"},
		{"~", home},
		{"$HOME", home},
		{"~" + user.Username, user.HomeDir},
		{"1 2", "1 2"},
	}

	for _, testItem := range testData {
		t.Run(testItem.source, func(t *testing.T) {
			result, err := ShellExpand(testItem.source)
			require.NoError(t, err)
			assert.Equal(t, testItem.expected, result)
		})
	}
}

func TestFindConfigurationIncludes(t *testing.T) {
	testID := fmt.Sprintf("%d", uint32(time.Now().UnixNano()))
	tempDir := os.TempDir()
	files := []string{
		filepath.Join(tempDir, "base."+testID+".conf"),
		filepath.Join(tempDir, "inc1."+testID+".conf"),
		filepath.Join(tempDir, "inc2."+testID+".conf"),
		filepath.Join(tempDir, "inc3."+testID+".conf"),
	}

	for _, file := range files {
		require.NoError(t, ioutil.WriteFile(file, []byte{}, fs.ModePerm))
		defer os.Remove(file) // defer stack is ok for cleanup
	}

	testData := []struct {
		includes []string
		expected []string
	}{
		// Invalid pattern
		{[]string{"[--]"}, nil},
		// Empty
		{[]string{"no-match"}, []string{}},
		// Existing files
		{files[2:4], files[2:4]},
		// GLOB patterns
		{[]string{"inc*." + testID + ".conf"}, files[1:]},
		{[]string{"*inc*." + testID + ".*"}, files[1:]},
		{[]string{"inc1." + testID + ".conf"}, files[1:2]},
		{[]string{"inc3." + testID + ".conf", "inc1." + testID + ".conf"}, []string{files[3], files[1]}},
		{[]string{"inc3." + testID + ".conf", "no-match"}, []string{files[3]}},
		// Does not include self
		{[]string{"base." + testID + ".conf"}, []string{}},
		{files[0:1], []string{}},
	}

	for _, test := range testData {
		t.Run(strings.Join(test.includes, ","), func(t *testing.T) {
			result, err := FindConfigurationIncludes(files[0], test.includes)
			if test.expected == nil {
				assert.Nil(t, result)
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
				if len(test.expected) == 0 {
					assert.Nil(t, result)
				} else {
					assert.Equal(t, test.expected, result)
				}
			}
		})
	}
}
