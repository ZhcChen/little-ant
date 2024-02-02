package main

type ComputerResources struct {
	Cpu             float64              `json:"cpu"`
	Memory          float64              `json:"memory"`
	Disk            float64              `json:"disk"`
	DiskSpeed       map[string]DiskSpeed `json:"diskSpeed"`
	NetworkUpload   uint64               `json:"networkUpload"`
	NetworkDownload uint64               `json:"networkDownload"`
	Ip              string               `json:"ip"`
}

type DiskSpeed struct {
	Name       string `json:"name"`
	ReadSpeed  uint64 `json:"readSpeed"`
	WriteSpeed uint64 `json:"writeSpeed"`
}

type WsRecord struct {
	Timestamp int64  `json:"timestamp"`
	AutoSend  bool   `json:"auto_send"`
	Key       string `json:"key"`
}

type WsMessage[T any] struct {
	Data T   `json:"data"`
	Type int `json:"type"`
}
