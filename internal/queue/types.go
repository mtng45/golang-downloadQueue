package queue

import (
	"context"
	"sync"
	"time"
)

type DownloadStatus string

const (
	StatusPending   DownloadStatus = "pending"
	StatusActive    DownloadStatus = "active"
	StatusPaused    DownloadStatus = "paused"
	StatusCompleted DownloadStatus = "completed"
	StatusError     DownloadStatus = "error"
	StatusCanceled  DownloadStatus = "canceled"
)

type DownloadItem struct {
	ID          string
	URL         string
	FilePath    string
	Status      DownloadStatus
	Progress    float64
	Error       error
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	BytesRead   int64
	TotalBytes  int64
	ctx         context.Context
	cancel      context.CancelFunc
}

type DownloadQueue struct {
	items         []*DownloadItem
	workQueue     chan *DownloadItem
	mu            sync.RWMutex // 排他制御（Read-Write Mutex）
	maxConcurrent int
}
