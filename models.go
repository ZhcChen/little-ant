package main

type ComputerResources struct {
	Cpu             float64
	Memory          float64
	Disk            float64
	DiskSpeed       map[string]DiskSpeed
	NetworkUpload   uint64
	NetworkDownload uint64
}

type DiskSpeed struct {
	Name       string
	ReadSpeed  uint64
	WriteSpeed uint64
}
