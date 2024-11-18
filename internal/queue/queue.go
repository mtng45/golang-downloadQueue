package queue

import (
	"context"
	"download-queue/internal/utils"
	"fmt"
	"time"
)

// アイテムをIDで検索
func (q *DownloadQueue) findItem(id string) *DownloadItem {
	for _, item := range q.items {
		if item.ID == id {
			return item
		}
	}
	return nil
}

// ダウンロードキューの初期化
func NewDownloadQueue(maxConcurrent int) *DownloadQueue {
	q := &DownloadQueue{
		items:         make([]*DownloadItem, 0),
		workQueue:     make(chan *DownloadItem, maxConcurrent),
		maxConcurrent: maxConcurrent,
	}

	// ワーカープールの開始
	go q.startWorkers()

	return q
}

// ダウンロードキューにアイテムを追加
func (q *DownloadQueue) Add(url, filePath string) (*DownloadItem, error) {
	item := &DownloadItem{
		ID:        utils.GenerateID(), // UUIDなどを生成
		URL:       url,
		FilePath:  filePath,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}

	q.mu.Lock()
	q.items = append(q.items, item)
	q.mu.Unlock()

	// ダウンロードの開始
	return item, q.Start(item.ID)
}

// ダウンロードの開始
func (q *DownloadQueue) Start(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	item := q.findItem(id)
	if item == nil {
		return fmt.Errorf("item not found: %s", id)
	}

	// ダウンロード開始できる状態でない場合はエラー
	if item.Status != StatusPending && item.Status != StatusPaused {
		return fmt.Errorf("cannot start download in status: %s", item.Status)
	}

	item.ctx, item.cancel = context.WithCancel(context.Background())
	item.Status = StatusPending
	q.workQueue <- item

	return nil
}

// ダウンロードの一時停止
func (q *DownloadQueue) Pause(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	item := q.findItem(id)
	if item == nil {
		return fmt.Errorf("item not found: %s", id)
	}

	if item.Status != StatusActive {
		return fmt.Errorf("cannot pause download in status: %s", item.Status)
	}

	item.cancel()
	item.Status = StatusPaused
	return nil
}

func (q *DownloadQueue) Cancel(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	item := q.findItem(id)
	if item == nil {
		return fmt.Errorf("item not found: %s", id)
	}

	if item.cancel != nil {
		item.cancel()
	}
	item.Status = StatusCanceled
	return nil
}

// ダウンロードの状態を取得
func (q *DownloadQueue) GetStatus() []*DownloadItem {
	q.mu.RLock()
	defer q.mu.RUnlock()

	// コピーを返してスレッドセーフにする
	result := make([]*DownloadItem, len(q.items))
	copy(result, q.items)
	return result
}

// ワーカープールの開始
func (q *DownloadQueue) startWorkers() {
	for item := range q.workQueue {
		go q.processDownload(item)
	}
}
