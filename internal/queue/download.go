package queue

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// 個々のダウンロードアイテムの処理
func (q *DownloadQueue) processDownload(item *DownloadItem) {
	// ダウンロード開始時の状態更新
	q.mu.Lock()                // 排他制御
	item.Status = StatusActive // ダウンロード中
	now := time.Now()          // 現在時刻
	item.StartedAt = &now      // 開始時刻
	q.mu.Unlock()              // 排他制御解除

	// ファイルの作成
	file, err := os.OpenFile(item.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		q.handleError(item, err)
		return
	}
	defer file.Close() // ファイルのクローズ

	// 現在のファイルサイズを取得（レジューム用）
	fileInfo, err := file.Stat()
	if err != nil {
		q.handleError(item, err)
		return
	}

	// HTTPリクエストの作成
	req, err := http.NewRequestWithContext(item.ctx, "GET", item.URL, nil)
	if err != nil {
		q.handleError(item, err)
		return
	}

	// レジューム用のヘッダー設定
	if fileInfo.Size() > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", fileInfo.Size()))
		item.BytesRead = fileInfo.Size()
	}

	// ダウンロードの実行
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		q.handleError(item, err)
		return
	}
	defer resp.Body.Close() // HTTPレスポンスのクローズ

	// トータルサイズの設定
	item.TotalBytes = resp.ContentLength + fileInfo.Size()

	// プログレス更新付きでダウンロード
	buffer := make([]byte, 32*1024)
	for {
		select {
		case <-item.ctx.Done(): // キャンセルされた場合
			return
		default:
			n, err := resp.Body.Read(buffer)
			if n > 0 {
				// ファイルへの書き込み
				_, writeErr := file.Write(buffer[:n])
				if writeErr != nil {
					q.handleError(item, writeErr)
					return
				}

				// ダウンロードしたバイト数の更新
				item.BytesRead += int64(n)
				item.Progress = float64(item.BytesRead) / float64(item.TotalBytes) * 100
			}

			if err == io.EOF {
				q.completeDownload(item) // ダウンロード完了
				return
			}

			if err != nil {
				q.handleError(item, err)
				return
			}
		}
	}
}

// エラー処理
func (q *DownloadQueue) handleError(item *DownloadItem, err error) {
	q.mu.Lock()         // 排他制御
	defer q.mu.Unlock() // 排他制御解除

	item.Status = StatusError // エラー
	item.Error = err
}

// ダウンロード完了
func (q *DownloadQueue) completeDownload(item *DownloadItem) {
	q.mu.Lock()         // 排他制御
	defer q.mu.Unlock() // 排他制御解除

	now := time.Now()             // 現在時刻
	item.CompletedAt = &now       // 完了時刻
	item.Status = StatusCompleted // 完了
	item.Progress = 100           // 100%
}
