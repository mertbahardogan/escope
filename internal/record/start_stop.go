package record

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/mertbahardogan/escope/internal/config"
	"github.com/mertbahardogan/escope/internal/connection"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type Session struct {
	PID     int    `yaml:"pid"`
	LogPath string `yaml:"log_path"`
}

func ValidateRecordIntervalSeconds(n int) error {
	if n < constants.RecordIntervalMinSeconds {
		return fmt.Errorf("record: --interval must be at least %d second(s)", constants.RecordIntervalMinSeconds)
	}
	if n > constants.RecordIntervalMaxSeconds {
		return fmt.Errorf("record: --interval cannot exceed %d seconds", constants.RecordIntervalMaxSeconds)
	}
	return nil
}

func resolveDesktopDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, constants.RecordDesktopDir)
	if st, err := os.Stat(dir); err == nil && st.IsDir() {
		return dir, nil
	}
	return "", fmt.Errorf("desktop directory not found: %s", dir)
}

func DefaultLogPath() (string, error) {
	dir, err := resolveDesktopDir()
	if err != nil {
		return "", err
	}
	name := constants.RecordLogFilenamePrefix + time.Now().Format(constants.RecordLogFilenameTimeLayout) + constants.RecordLogFileExtension
	return filepath.Join(dir, name), nil
}

func sessionFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, constants.RecordSessionFileName), nil
}

func readSession() (*Session, error) {
	path, err := sessionFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(constants.ErrRecordNoActiveSession)
		}
		return nil, err
	}
	var s Session
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	if s.PID == 0 || s.LogPath == constants.EmptyString {
		return nil, errors.New(constants.ErrRecordInvalidSession)
	}
	return &s, nil
}

func writeSession(s *Session) error {
	path, err := sessionFilePath()
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func clearSession() error {
	path, err := sessionFilePath()
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func clusterLogIdentity(cmd *cobra.Command) (alias string, host string) {
	host = connection.CurrentHost()
	r := cmd.Root()
	if v, _ := r.PersistentFlags().GetString("alias"); v != constants.EmptyString {
		return v, host
	}
	if ah, err := config.GetActiveHost(); err == nil && ah != constants.EmptyString {
		return ah, host
	}
	return constants.EmptyString, host
}

func daemonChildPersistentArgs(cmd *cobra.Command) []string {
	r := cmd.Root()
	var out []string
	if v, _ := r.PersistentFlags().GetString("host"); v != constants.EmptyString {
		out = append(out, "-H", v)
	}
	if v, _ := r.PersistentFlags().GetString("username"); v != constants.EmptyString {
		out = append(out, "-u", v)
	}
	if v, _ := r.PersistentFlags().GetString("password"); v != constants.EmptyString {
		out = append(out, "-p", v)
	}
	if v, _ := r.PersistentFlags().GetBool("secure"); v {
		out = append(out, "--secure")
	}
	if v, _ := r.PersistentFlags().GetString("alias"); v != constants.EmptyString {
		out = append(out, "-a", v)
	}
	return out
}

func StartRecording(cmd *cobra.Command) error {
	if s, err := readSession(); err == nil {
		if processAlive(s.PID) {
			return fmt.Errorf("record is already running (pid %d); run `escope record stop` before starting again", s.PID)
		}
		_ = clearSession()
	}
	if err := connection.ApplyPersistentConnection(cmd); err != nil {
		return err
	}
	intervalSec, err := cmd.Flags().GetInt("interval")
	if err != nil {
		return err
	}
	if err := ValidateRecordIntervalSeconds(intervalSec); err != nil {
		return err
	}
	logPath, err := DefaultLogPath()
	if err != nil {
		return err
	}
	absPath, err := filepath.Abs(logPath)
	if err != nil {
		return err
	}
	f, err := os.Create(logPath)
	if err != nil {
		return err
	}
	alias, _ := clusterLogIdentity(cmd)
	if err := writeLogHeader(f, alias, intervalSec); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	child := append([]string{}, os.Args[0])
	child = append(child, daemonChildPersistentArgs(cmd)...)
	child = append(child, "record", "daemon", "--log", logPath, "--interval", strconv.Itoa(intervalSec))
	c := exec.Command(child[0], child[1:]...)
	c.Stdin = nil
	c.Stdout = nil
	c.Stderr = nil
	prepareDetachedChild(c)
	if err := c.Start(); err != nil {
		return err
	}
	pid := c.Process.Pid
	_ = c.Process.Release()
	if err := writeSession(&Session{PID: pid, LogPath: logPath}); err != nil {
		return err
	}
	fmt.Printf("Recording to %s (pid %d)\n", absPath, pid)
	return nil
}

func StopRecording() error {
	s, err := readSession()
	if err != nil {
		return err
	}
	if proc, err := os.FindProcess(s.PID); err == nil {
		_ = proc.Signal(syscall.SIGTERM)
	}
	_ = clearSession()
	return nil
}
