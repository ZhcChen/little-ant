package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"log"
	"net/http"
	"time"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var wsConnMap = make(map[string]*websocket.Conn)
var kkk = ""
var computerResources *ComputerResources = &ComputerResources{
	Cpu:             0,
	Memory:          0,
	Disk:            0,
	DiskSpeed:       map[string]DiskSpeed{},
	NetworkUpload:   0,
	NetworkDownload: 0,
}

func Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("长链接失败", err)
		return
	}

	if conn != nil {
		kkk = r.Header.Get("Sec-Websocket-Key")
		wsConnMap[kkk] = conn
	}

	go func() {
		for {
			time.Sleep(time.Second * 1)
			_conn := wsConnMap[kkk]
			if _conn != nil {
				computerResourcesJson, err := json.Marshal(computerResources)
				if err != nil {
					continue
				}
				_conn.WriteMessage(1, computerResourcesJson)
			}
		}
	}()

	for {
		t, msg, readMessageErr := conn.ReadMessage()
		if readMessageErr != nil {
			break
		}
		writeMessageErr := conn.WriteMessage(t, msg)
		if writeMessageErr != nil {
			return
		}
	}
}

// Ws websocket
func Ws(c *gin.Context) {
	wsHandler(c.Writer, c.Request)
}

// MonitorComputerResources 这里监控计算机本身资源
func MonitorComputerResources() {
	go func() {
		log.Println("开始监控")

		for {
			time.Sleep(time.Second * 1)

			// 获取CPU使用率
			cpuPercent, _ := cpu.Percent(0, false)
			computerResources.Cpu = cpuPercent[0]
			fmt.Printf("CPU Usage: %.2f%%\n", cpuPercent[0])

			// 获取内存使用率
			memInfo, _ := mem.VirtualMemory()
			computerResources.Memory = memInfo.UsedPercent
			fmt.Printf("Memory Usage: %.2f%%\n", memInfo.UsedPercent)

			// 获取磁盘使用率
			diskInfo, _ := disk.Usage("/")
			computerResources.Disk = diskInfo.UsedPercent
			fmt.Printf("Disk Usage: %.2f%%\n", diskInfo.UsedPercent)

			// 获取磁盘读写速率
			DiskIoRwSpeed()

			// 获取网络信息
			netInfo, _ := net.IOCounters(false)
			computerResources.NetworkUpload = netInfo[0].BytesSent
			computerResources.NetworkDownload = netInfo[0].BytesRecv
			fmt.Printf("Network Upload: %d bytes\n", netInfo[0].BytesSent)
			fmt.Printf("Network Download: %d bytes\n", netInfo[0].BytesRecv)
		}
	}()
}

func DiskIoRwSpeed() {
	// 获取磁盘IO统计信息
	before, err := disk.IOCounters()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for k, _ := range before {
		computerResources.DiskSpeed[k] = DiskSpeed{
			Name:       k,
			ReadSpeed:  0,
			WriteSpeed: 0,
		}
	}

	// 等待一段时间，例如1秒
	time.Sleep(1 * time.Second)

	// 再次获取磁盘IO统计信息
	after, err := disk.IOCounters()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 计算磁盘读写速率
	for k, v := range after {
		computerResources.DiskSpeed[k] = DiskSpeed{
			Name:       k,
			ReadSpeed:  v.ReadBytes - before[k].ReadBytes,
			WriteSpeed: v.WriteBytes - before[k].WriteBytes,
		}
	}
}
