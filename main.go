package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"sync"
)

// IndexedChecksum は、行のインデックスとそのSHA-256チェックサムを保持する構造体です。
type IndexedChecksum struct {
	Index    int    // 行のインデックス
	Checksum string // 計算されたSHA-256チェックサム
}

// computeChecksum は、指定された行に対してSHA-256チェックサムを計算し、結果をチャネルに送信します。
func computeChecksum(line string, index int, wg *sync.WaitGroup, out chan<- IndexedChecksum) {
	defer wg.Done()                                          // 作業完了時にWaitGroupのカウンターをデクリメント
	hash := sha256.Sum256([]byte(line))                      // SHA-256ハッシュを計算
	checksum := hex.EncodeToString(hash[:])                  // ハッシュ値をHEXダンプに変換
	out <- IndexedChecksum{Index: index, Checksum: checksum} // 結果をチャネルに送信
}

func main() {
	// コマンドライン引数のチェック
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <filename>")
		return
	}

	filename := os.Args[1]
	file, err := os.Open(filename) // ファイルを開く
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err)
		return
	}
	defer file.Close() // 関数終了時にファイルを閉じる

	scanner := bufio.NewScanner(file)
	var wg sync.WaitGroup
	checksumChan := make(chan IndexedChecksum)

	// ファイルの各行を並行して処理
	for index := 0; scanner.Scan(); index++ {
		wg.Add(1)
		go computeChecksum(scanner.Text(), index, &wg, checksumChan)
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		return
	}

	// 全てのゴルーチンが完了するのを待つゴルーチン
	go func() {
		wg.Wait()           // すべてのゴルーチンが完了するのを待つ
		close(checksumChan) // チャネルをクローズ
	}()

	// チェックサム結果を収集し、インデックスに従ってソート
	var results []IndexedChecksum
	for checksum := range checksumChan {
		results = append(results, checksum)
	}

	// 結果をインデックス順にソート
	sort.Slice(results, func(i, j int) bool {
		return results[i].Index < results[j].Index
	})

	// 結果を表示
	for _, res := range results {
		fmt.Println(res.Checksum)
	}
}
