package main

import (
	"download-queue/internal/queue"
	"fmt"
	"time"
)

func main() {
	// 最大3つの同時ダウンロードを許可
	downloadQueue := queue.NewDownloadQueue(3)

	// ダウンロードの追加
	// Go のリリースバイナリなど
	// テスト用のURLに変更
	item1, _ := downloadQueue.Add("https://httpbin.org/bytes/10485760", "./downloads/file1.bin") // 10MB
	item2, _ := downloadQueue.Add("https://httpbin.org/bytes/5242880", "./downloads/file2.bin")  // 5MB

	// ステータス監視
	go func() {
		for {
			status := downloadQueue.GetStatus()
			for _, item := range status {
				fmt.Printf("ID: %s, Status: %s, Progress: %.2f%%\n",
					item.ID, item.Status, item.Progress)
			}
			time.Sleep(time.Second)
		}
	}()

	// 操作例
	time.Sleep(2 * time.Second)
	downloadQueue.Pause(item1.ID) // 一時停止

	time.Sleep(2 * time.Second)
	downloadQueue.Start(item1.ID) // 再開

	time.Sleep(2 * time.Second)
	downloadQueue.Cancel(item2.ID) // キャンセル

	// メインプログラムが終了しないように待機
	select {}
}
