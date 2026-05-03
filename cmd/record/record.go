package record

import (
	"fmt"

	"github.com/mertbahardogan/escope/cmd/core"
	"github.com/mertbahardogan/escope/internal/connection"
	"github.com/mertbahardogan/escope/internal/constants"
	intrec "github.com/mertbahardogan/escope/internal/record"
	"github.com/spf13/cobra"
)

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record cluster hot threads, shard activity, and node metrics to a file",
	Long: "Runs a background sampler that appends timed snapshots to a log file on the desktop (~/Desktop).\n\n" +
		"Session state and the sampler process id are stored in a small session file, not in the host config sessions map.\n\n" +
		"Use the same global connection flags as other commands when starting. Sampling uses `--interval` in seconds (default from constants).\n\n" +
		"Examples:\n" +
		"  escope record start\n" +
		"  escope record start --interval 120\n" +
		"  escope record stop",
	Example: `  escope record start
  escope record stop`,
}

var recordStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start background recording to a desktop log file",
	RunE: func(cmd *cobra.Command, args []string) error {
		return intrec.StartRecording(cmd)
	},
}

var recordStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop background recording",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := intrec.StopRecording(); err != nil {
			return err
		}
		fmt.Println("Stop sent to the record sampler; it should finish and close the log shortly.")
		return nil
	},
}

var recordDaemonCmd = &cobra.Command{
	Use:    "daemon",
	Short:  "Internal sampler process (do not run directly)",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := connection.ApplyPersistentConnection(cmd); err != nil {
			return err
		}
		logPath, err := cmd.Flags().GetString("log")
		if err != nil {
			return err
		}
		if logPath == "" {
			return fmt.Errorf("--log is required")
		}
		intervalSec, err := cmd.Flags().GetInt("interval")
		if err != nil {
			return err
		}
		if err := intrec.ValidateRecordIntervalSeconds(intervalSec); err != nil {
			return err
		}
		return intrec.RunDaemon(logPath, intervalSec)
	},
}

func init() {
	recordStartCmd.Flags().Int("interval", constants.RecordIntervalSeconds, "Sample interval in seconds between ticks")
	recordDaemonCmd.Flags().String("log", "", "Log file path")
	recordDaemonCmd.Flags().Int("interval", constants.RecordIntervalSeconds, "Sample interval in seconds")
	_ = recordDaemonCmd.MarkFlagRequired("log")
	recordCmd.AddCommand(recordStartCmd)
	recordCmd.AddCommand(recordStopCmd)
	recordCmd.AddCommand(recordDaemonCmd)
	core.RootCmd.AddCommand(recordCmd)
}
