package main

import "log"

var (
	Version    string
	CommitHash string
	BuildDate  string
)

func main() {
	// 你可以在日誌或啟動信息中打印這些版本信息
	log.Printf("Starting server version: %s, commit: %s, built at: %s", Version, CommitHash, BuildDate)
	// ...
}
