package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strconv"
)

type Configuration struct {
	Job    map[string]string `json:"job"`
	Cookie string            `json:"cookie"`
}

func convertJobKeyType(job map[string]string) map[int]string {
	result := make(map[int]string)
	for key, value := range job {
		newKey, err := strconv.Atoi(key)
		if err != nil {
			log.Fatalf("无法转换键: %v", err)
		}
		result[newKey] = value
	}
	return result
}

// Load config from file
func LoadConfig(configFile string) (*Configuration, error) {
	config := &Configuration{}
	file, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("无法打开文件: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("无法关闭文件: %v", err)
		}
	}(file)
	bytes, err := io.ReadAll(file)
	if err := json.Unmarshal(bytes, &config); err != nil {
		log.Fatalf("无法解析 JSON 数据: %v", err)
	}
	return config, err
}
