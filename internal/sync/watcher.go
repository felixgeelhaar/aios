package sync

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type WatchEvent struct {
	Path string
	When time.Time
}

type Watcher interface {
	Events() <-chan WatchEvent
}

type PollingWatcher struct {
	interval time.Duration
}

func NewPollingWatcher(interval time.Duration) *PollingWatcher {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	return &PollingWatcher{interval: interval}
}

func (w *PollingWatcher) Watch(ctx context.Context, engine *Engine, paths []string, repair func(string) error) (<-chan WatchEvent, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("paths are required")
	}

	stamps := make(map[string]string, len(paths))
	for _, p := range paths {
		stamp, err := pathStamp(p)
		if err != nil {
			return nil, err
		}
		stamps[p] = stamp
	}

	events := make(chan WatchEvent, len(paths))
	ticker := time.NewTicker(w.interval)
	go func() {
		defer ticker.Stop()
		defer close(events)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for _, p := range paths {
					stamp, err := pathStamp(p)
					if err != nil {
						continue
					}
					if stamp == stamps[p] {
						continue
					}
					stamps[p] = stamp
					ev := WatchEvent{Path: p, When: time.Now()}
					select {
					case events <- ev:
					default:
					}

					if engine != nil {
						engine.MarkDrifted()
						engine.MarkRepairing()
					}
					if repair == nil {
						continue
					}
					if err := repair(p); err != nil {
						if engine != nil {
							engine.MarkDrifted()
						}
						continue
					}
					if engine != nil {
						engine.MarkStable()
					}
				}
			}
		}
	}()

	return events, nil
}

func pathStamp(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return fmt.Sprintf("f:%d:%d", info.ModTime().UnixNano(), info.Size()), nil
	}
	stamp := ""
	err = filepath.WalkDir(path, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		stamp += fmt.Sprintf("%s:%d:%d|", p, fi.ModTime().UnixNano(), fi.Size())
		return nil
	})
	if err != nil && err != io.EOF {
		return "", err
	}
	return stamp, nil
}
