package main

import (
	"github.com/joho/godotenv"
)

const (
	Version   = "1.0.0-alpha"
	DebugMode = false
)

func loadEnv() {
	// 環境変数 読み込み
	// ref.) https://zenn.dev/a_ichi1/articles/c9f3870350c5e2
	err := godotenv.Load(".env")
	if err != nil {
		logger.Println("環境変数の読み込みに失敗しました")
		logger.Fatal(err)
	}
}
