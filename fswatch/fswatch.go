package fswatch

import (
	"gorond/util"
	"os"
	"path/filepath"
	"time"
)

type Event struct {
	Added    []string
	Removed  []string
	Modified []string
	Error    error
}

// デフォルトの監視間隔
var WatchInterval = time.Second * 10

// イベントコールバック関数
type EventCallback func(event Event)

func (self *Event) HasModifiedFiles() bool {
	if 0 < len(self.Added) || 0 < len(self.Removed) || 0 < len(self.Modified) {
		return true
	}
	return false
}

type Watcher struct {
	EventChan   chan Event
	TimeoutChan chan Event
	exitChan    chan int

	// check interval.
	interval time.Duration

	// path -> pattern
	paths map[string]string

	// path -> timestamp
	context map[string]int64

	running bool
}

func NewWatcher(interval time.Duration) (*Watcher, error) {
	watcher := &Watcher{}
	watcher.interval = interval
	watcher.paths = map[string]string{}
	watcher.context = map[string]int64{}
	return watcher, nil
}

func (self *Watcher) Add(path string) {
	self.paths[path] = path
}

func (self *Watcher) AddDir(path string, pattern string) {
	self.paths[path] = pattern
}

func (self *Watcher) Start(enableTimeout bool) error {
	self.EventChan = make(chan Event)
	self.TimeoutChan = make(chan Event)
	self.exitChan = make(chan int)

	// 初期化処理として、一度ファイルシステムのチェックを実行
	_, err := self.watchFileSystem()
	if err != nil {
		return err
	}

	// イベントハンドラ
	go func() {

	EVENT_LOOP:
		for {
			select {
			case <-self.exitChan:
				self.running = false
				break EVENT_LOOP
			case <-time.After(self.interval):
				// watch file system
				if self.running {
					event, err := self.watchFileSystem()
					if err != nil {
						event.Error = err
						self.EventChan <- event
					} else if event.HasModifiedFiles() {
						self.EventChan <- event
					} else if enableTimeout {
						self.TimeoutChan <- event
					}
				} else {
					self.exitChan <- 0
				}
			}
		}

	}()

	self.running = true

	return nil
}

func (self *Watcher) Close() {
	close(self.EventChan)
	close(self.TimeoutChan)

	self.exitChan <- 0

	close(self.exitChan)
}

func (self *Watcher) watchFileSystem() (Event, error) {
	event := Event{}

	watchFiles := []string{}
	for path, pattern := range self.paths {
		stat, err := os.Stat(path)
		if err != nil {
			return event, err
		}

		if stat.Mode().IsDir() {
			files, err := util.FileList(path, pattern)
			if err != nil {
				return event, err
			}

			for _, file := range files {
				filePath := filepath.Join(path, file)
				watchFiles = append(watchFiles, filePath)
			}
		} else if stat.Mode().IsRegular() {
			watchFiles = append(watchFiles, path)
		}
	}

	// 変更点を調査
	modifiedFiles := []string{}
	addedFiles := []string{}
	removedFiles := []string{}
	timestamps := map[string]int64{}
	for _, file := range watchFiles {
		stat, err := os.Stat(file)
		if err != nil {
			return event, err
		}

		// タイムスタンプを過去と比較
		ts, exists := self.context[file]
		if !exists {
			// 前回のチェック時に存在しなかった場合は追加として扱う
			addedFiles = append(addedFiles, file)
		} else if ts != stat.ModTime().Unix() {
			// タイムスタンプが異なる場合は変更として扱う
			modifiedFiles = append(modifiedFiles, file)
		}

		timestamps[file] = stat.ModTime().Unix()
	}

	// 削除されたファイルを検知する
	for file, _ := range self.context {
		_, exists := timestamps[file]
		if !exists {
			removedFiles = append(removedFiles, file)
		}
	}

	// コンテキストを更新する
	self.context = timestamps

	// イベントに設定する
	event.Added = addedFiles
	event.Modified = modifiedFiles
	event.Removed = removedFiles

	return event, nil
}

// 基本的なファイル監視ルーチンを実行する
//
// paths: []監視ファイル
// dirs: map[監視ディレクトリ]ファイル名パターン
// configChanged: イベントコールバック
func StartWatcher(paths []string, dirs map[string]string, configChanged EventCallback) (*Watcher, error) {
	watcher, err := NewWatcher(WatchInterval)
	if err != nil {
		return nil, err
	}

	go func() {
	WATCH_LOOP:
		for {
			select {
			case event := <-watcher.EventChan:
				if event.HasModifiedFiles() {
					configChanged(event)
				}
			case <-time.After(time.Second):
				if !watcher.running {
					break WATCH_LOOP
				}
			}
		}
	}()

	for _, path := range paths {
		watcher.Add(path)
	}
	for dir, pattern := range dirs {
		watcher.AddDir(dir, pattern)
	}
	err = watcher.Start(false)
	if err != nil {
		return nil, err
	}

	return watcher, nil
}
