package fswatch

import (
	"os"
	"testing"
	"time"
)

var INTERVAL = time.Second

// イベントなしテスト
func TestWatchNoEvent(t *testing.T) {
	watcher := startWatcher(t)
	defer watcher.Close()

	select {
	case event := <-watcher.EventChan:
		if event.Error != nil {
			t.Fatal(event.Error)
		}

		// 通知がきたら失敗
		t.Errorf("(expected: -) added: %d, modified: %d, removed: %d", event.Added, event.Modified, event.Removed)
	case event := <-watcher.TimeoutChan:
		if event.Error != nil {
			t.Fatal(event.Error)
		}

		if event.HasModifiedFiles() {
			t.Errorf("(expected: 0) added: %d, modified: %d, removed: %d", event.Added, event.Modified, event.Removed)
		}
	}
}

// イベント通知テスト
func TestNotifyEvent(t *testing.T) {
	watcher := startWatcher(t)
	defer watcher.Close()

	createTestFile(t, ".notify.conf")

	select {
	case event := <-watcher.EventChan:
		if event.Error != nil {
			t.Fatal(event.Error)
		}

		if 1 != len(event.Added) {
			t.Errorf("(expected: added: 1) added: %v, modified: %v, removed: %v", event.Added, event.Modified, event.Removed)
		}
	case event := <-watcher.TimeoutChan:
		if event.Error != nil {
			t.Fatal(event.Error)
		}

		// タイムアウトしたら失敗
		t.Errorf("(expected: -) added: %v, modified: %v, removed: %v", event.Added, event.Modified, event.Removed)
	}

	removeTestFile(t, ".notify.conf")
}

// 追加イベント検知テスト
func TestWatchAddEvent(t *testing.T) {
	watcher := createWatcher(t)

	createTestFile(t, ".add.conf")

	event, err := watcher.watchFileSystem()

	if err != nil {
		t.Fatal(err)
	}

	if 1 != len(event.Added) {
		t.Errorf("(expected: added: 1) added: %v, modified: %v, removed: %v", event.Added, event.Modified, event.Removed)
	}

	removeTestFile(t, ".add.conf")
}

// 変更イベント検知テスト
func TestWatchModifyEvent(t *testing.T) {
	createTestFile(t, ".modify.conf")

	watcher := createWatcher(t)

	// 変更後のタイムスタンプをずらす
	time.Sleep(time.Second + time.Second/10)
	modifyTestFile(t, ".modify.conf", "test")

	event, err := watcher.watchFileSystem()
	if err != nil {
		t.Fatal(err)
	}
	if 1 != len(event.Modified) {
		t.Errorf("(expected: modified: 1) added: %v, modified: %v, removed: %v", event.Added, event.Modified, event.Removed)
	}

	removeTestFile(t, ".modify.conf")
}

// 削除イベント検知テスト
func TestWatchRemoveEvent(t *testing.T) {
	createTestFile(t, ".remove.conf")

	watcher := createWatcher(t)

	os.Remove(".remove.conf")

	event, err := watcher.watchFileSystem()
	if err != nil {
		t.Fatal(err)
	}
	if 1 != len(event.Removed) {
		t.Errorf("(expected: removed: 1) added: %v, modified: %v, removed: %v", event.Added, event.Modified, event.Removed)
	}
}

// 基本監視のテスト
func TestStartWatcher(t *testing.T) {
	createTestFile(t, ".startwatch.conf")

	rc := make(chan int)

	paths := []string{}
	dirs := map[string]string{"./": `.+\.conf`}
	WatchInterval = time.Second
	watcher, err := StartWatcher(paths, dirs, func(event Event) {
		rc <- 0
	})
	if err != nil {
		t.Error(err)
	}
	defer watcher.Close()

	// 1秒待って変更イベントをテスト
	time.Sleep(time.Second)
	modifyTestFile(t, ".startwatch.conf", "modified")

	if code := <-rc; code != 0 {
		t.Errorf("watch event result: %v", code)
	}

	// 1秒待って削除イベントをテスト
	time.Sleep(time.Second)
	removeTestFile(t, ".startwatch.conf")

	if code := <-rc; code != 0 {
		t.Errorf("watch event result: %v", code)
	}
}

// 以下テスト用ヘルパー ------------------------------------
func createWatcher(t *testing.T) *Watcher {
	watcher, err := NewWatcher(INTERVAL)
	if err != nil {
		t.Fatal(err)
	}

	watcher.AddDir("./", `.+\.conf`)

	_, err = watcher.watchFileSystem()
	if err != nil {
		t.Fatal(err)
	}

	return watcher
}

func startWatcher(t *testing.T) *Watcher {
	watcher, err := NewWatcher(INTERVAL)
	if err != nil {
		t.Fatal(err)
	}

	watcher.AddDir("./", `.+\.conf`)

	err = watcher.Start(true)
	if err != nil {
		t.Fatal(err)
	}

	return watcher
}

func createTestFile(t *testing.T, name string) {
	f, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func modifyTestFile(t *testing.T, name string, content string) {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0666)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte(content))
	if err != nil {
		t.Fatal(err)
	}
	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func removeTestFile(t *testing.T, name string) {
	err := os.Remove(name)
	if err != nil {
		t.Fatal(err)
	}
}
