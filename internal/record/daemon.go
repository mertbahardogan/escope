package record

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/mertbahardogan/escope/internal/config"
	"github.com/mertbahardogan/escope/internal/connection"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/elastic"
	"github.com/mertbahardogan/escope/internal/util"
)

func blockSeparator() string {
	return strings.Repeat("=", constants.RecordBlockSeparatorWidth)
}

func formatLocalTimestamp(t time.Time) string {
	return t.In(time.Local).Format("2006-01-02 15:04:05 MST")
}

func recordSamplerBannerLine(intervalSec int) string {
	return fmt.Sprintf(constants.RecordLogPreambleFmt,
		intervalSec,
		constants.RecordSectionHotThreads,
		constants.RecordSectionShardActivity,
		constants.RecordSectionNodeActivity,
	)
}

func writeLogHeader(w io.Writer, clusterAlias string, intervalSec int) error {
	ver, _ := config.GetInstalledVersion()
	if ver == constants.EmptyString {
		ver = constants.MsgRecordVersionUnknown
	}
	clusterLabel := clusterAlias
	if clusterLabel == constants.EmptyString {
		clusterLabel = constants.MsgRecordClusterAliasNone
	}
	if _, err := fmt.Fprintf(w, "# ESCOPE Record Log #\nescope_version: %s\ncluster: %s\nstarted_local: %s\n",
		ver, clusterLabel, formatLocalTimestamp(time.Now())); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, recordSamplerBannerLine(intervalSec)); err != nil {
		return err
	}
	return nil
}

func formatRecordSampleError(section string, err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Sprintf(constants.RecordSampleErrDeadlineFmt, section, err)
	}
	return fmt.Sprintf(constants.RecordSampleErrGenericFmt, section, err)
}

func writeTickHeader(w io.Writer, t time.Time) error {
	_, err := fmt.Fprintf(w, "\n%s\nTICK %s\n%s\n",
		blockSeparator(),
		formatLocalTimestamp(t),
		blockSeparator())
	return err
}

func writeSection(w io.Writer, title, body string) error {
	_, err := fmt.Fprintf(w, "--- %s ---\n%s\n", title, body)
	return err
}

func appendLogFooter(logPath string) error {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "\n%s\nstopped_local: %s\n%s\n",
		blockSeparator(),
		formatLocalTimestamp(time.Now()),
		blockSeparator())
	return err
}

type shardPlace struct {
	state string
	node  string
}

func fetchHotThreadsText(ctx context.Context, client *elasticsearch.Client) (string, error) {
	res, err := client.Nodes.HotThreads(
		client.Nodes.HotThreads.WithContext(ctx),
		client.Nodes.HotThreads.WithTimeout(time.Duration(constants.RecordHotThreadsServerTimeoutSecs)*time.Second),
		client.Nodes.HotThreads.WithThreads(constants.RecordHotThreadsThreadCount),
		client.Nodes.HotThreads.WithSnapshots(constants.RecordHotThreadsSnapshotCount),
		client.Nodes.HotThreads.WithInterval(time.Duration(constants.RecordHotThreadsIntervalMillis)*time.Millisecond),
		client.Nodes.HotThreads.WithIgnoreIdleThreads(true),
	)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.IsError() {
		b, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("hot_threads: %s: %s", res.Status(), string(b))
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func fetchShardActivity(ctx context.Context, client *elasticsearch.Client, prev map[string]shardPlace) (string, map[string]shardPlace, error) {
	w := elastic.NewClientWrapper(client)
	data, err := w.GetShards(ctx)
	if err != nil {
		return "", prev, err
	}
	shards, ok := data[constants.EmptyString].([]map[string]interface{})
	if !ok {
		return "", prev, fmt.Errorf("unexpected shards response shape")
	}

	counts := map[string]int{}
	next := make(map[string]shardPlace, len(shards))
	var b strings.Builder
	for _, s := range shards {
		st := util.GetStringField(s, constants.StateField)
		if st != constants.EmptyString {
			counts[st]++
		}
		idx := util.GetStringField(s, constants.IndexField)
		sh := util.GetStringField(s, constants.ShardField)
		pr := util.GetStringField(s, constants.PrirepField)
		node := util.GetStringField(s, constants.ShardRowNodeField)
		key := fmt.Sprintf(constants.RecordShardKeyFmt, idx, sh, pr)
		next[key] = shardPlace{state: st, node: node}
	}

	fmt.Fprintf(&b, "summary relocating=%d initializing=%d unassigned=%d started=%d\n",
		counts[constants.ShardStateRelocating],
		counts[constants.ShardStateInitializing],
		counts[constants.ShardStateUnassigned],
		counts[constants.ShardStateStarted])

	if prev == nil {
		fmt.Fprintf(&b, "assignment_changes: (baseline — deltas on next sample)\n")
		return b.String(), next, nil
	}

	fmt.Fprintf(&b, "assignment_changes:\n")
	changed := 0
	for k, cur := range next {
		old, ok := prev[k]
		if !ok || old.state != cur.state || old.node != cur.node {
			if ok {
				fmt.Fprintf(&b, "  %s: %s@%s -> %s@%s\n", k, old.state, old.node, cur.state, cur.node)
			} else {
				fmt.Fprintf(&b, "  %s: (new) %s@%s\n", k, cur.state, cur.node)
			}
			changed++
		}
	}
	for k, old := range prev {
		if _, ok := next[k]; !ok {
			fmt.Fprintf(&b, "  %s: removed (was %s@%s)\n", k, old.state, old.node)
			changed++
		}
	}
	if changed == 0 {
		fmt.Fprintf(&b, "  (no changes)\n")
	}
	return b.String(), next, nil
}

func fetchNodeActivity(ctx context.Context, client *elasticsearch.Client) (string, error) {
	res, err := client.Cat.Nodes(
		client.Cat.Nodes.WithContext(ctx),
		client.Cat.Nodes.WithFormat("json"),
		client.Cat.Nodes.WithH("name", "heap.percent", "ram.percent", "cpu", "load_1m", "node.role", "disk.used_percent"),
	)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.IsError() {
		b, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("cat nodes: %s: %s", res.Status(), string(b))
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func sampleTickContextDuration() time.Duration {
	d := time.Duration(constants.RecordTickTimeoutSeconds) * time.Second
	if n, err := config.GetConnectionTimeout(); err == nil && n > 0 {
		cfg := time.Duration(n) * time.Second
		if cfg > d {
			d = cfg
		}
	}
	return d
}

func RunDaemon(logPath string, intervalSec int) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	client := connection.NewRecordSamplerClient()
	if client == nil {
		return errors.New(constants.ErrRecordClientUnavailable)
	}

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	tickDur := time.Duration(intervalSec) * time.Second
	ticker := time.NewTicker(tickDur)
	defer ticker.Stop()

	var prev map[string]shardPlace

	sample := func() error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		tickCtx, c2 := context.WithTimeout(context.Background(), sampleTickContextDuration())
		defer c2()
		t := time.Now()
		if err := writeTickHeader(f, t); err != nil {
			return err
		}
		ht, err := fetchHotThreadsText(tickCtx, client)
		if err != nil {
			ht = formatRecordSampleError(constants.RecordSectionHotThreads, err)
		}
		if err := writeSection(f, constants.RecordSectionHotThreads, ht); err != nil {
			return err
		}
		shardText, next, err := fetchShardActivity(tickCtx, client, prev)
		if err != nil {
			shardText = formatRecordSampleError(constants.RecordSectionShardActivity, err)
			next = prev
		}
		prev = next
		if err := writeSection(f, constants.RecordSectionShardActivity, shardText); err != nil {
			return err
		}
		nodeText, err := fetchNodeActivity(tickCtx, client)
		if err != nil {
			nodeText = formatRecordSampleError(constants.RecordSectionNodeActivity, err)
		}
		if err := writeSection(f, constants.RecordSectionNodeActivity, nodeText); err != nil {
			return err
		}
		return f.Sync()
	}

	if err := sample(); err != nil {
		if ctx.Err() != nil {
			return appendLogFooter(logPath)
		}
		fmt.Fprintf(f, constants.RecordErrFirstSampleFmt+"\n", err)
		_ = f.Sync()
	}

	for {
		select {
		case <-ctx.Done():
			return appendLogFooter(logPath)
		case <-ticker.C:
			if err := sample(); err != nil {
				if ctx.Err() != nil {
					return appendLogFooter(logPath)
				}
				fmt.Fprintf(f, constants.RecordErrLaterSampleFmt+"\n", err)
				_ = f.Sync()
			}
		}
	}
}
