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
	"strconv"
	"time"
)

const PingTimeout = 10

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var wsConnMap = make(map[string]*websocket.Conn)
var wsRecordMap = make(map[string]*WsRecord)
var computerResources = &ComputerResources{
	Cpu:             0,
	Memory:          0,
	Disk:            0,
	DiskSpeed:       map[string]DiskSpeed{},
	NetworkUpload:   0,
	NetworkDownload: 0,
	Ip:              "",
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

	key := ""
	if conn != nil {
		key = r.Header.Get("Sec-Websocket-Key")
		wsConnMap[key] = conn
		wsRecordMap[key] = &WsRecord{
			Timestamp: time.Now().Unix(),
			AutoSend:  true,
			Key:       key,
		}
	}

	// 推机器监控数据
	go func() {
		_conn := wsConnMap[key]
		wsRecord := wsRecordMap[key]

		defer _conn.Close()

		for {
			if _conn != nil {
				// 移除超时的长链接
				if (wsRecord.Timestamp + PingTimeout) < time.Now().Unix() {
					delete(wsConnMap, key)
					break
				}
				wsMessage := WsMessage[ComputerResources]{
					Data: *computerResources,
					Type: 1,
				}
				wsMessageJson, wsMessageErr := json.Marshal(wsMessage)
				if wsMessageErr != nil {
					continue
				}
				connWriteMessage := _conn.WriteMessage(1, wsMessageJson)
				if connWriteMessage != nil {
					log.Println("推机器监控数据异常：", connWriteMessage)
				}
			} else {
				break
			}
			time.Sleep(time.Second)
		}
	}()

	// 消息接收
	for {
		t, msg, readMessageErr := conn.ReadMessage()
		if readMessageErr != nil {
			break
		}
		var data WsMessage[any]
		if json.Unmarshal(msg, &data) != nil {
			continue
		}
		// 心跳包
		if data.Type == 0 && data.Data.(string) == "ping" {
			ackMsg := WsMessage[string]{
				Data: "pong",
				Type: 0,
			}
			ackMsgBytes, ackMsgBytesErr := json.Marshal(ackMsg)
			if ackMsgBytesErr != nil {
				log.Println("心跳包数据json异常：", ackMsgBytesErr)
				continue
			}
			writeMessageErr := conn.WriteMessage(t, ackMsgBytes)
			if writeMessageErr != nil {
				log.Println("心跳包响应推送异常：", writeMessageErr)
				continue
			}
			wsRecordMap[key].Timestamp = time.Now().Unix()
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
		for {
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

			// 获取IP
			GetRealIp()

			time.Sleep(time.Second * 1)
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

func GetRealIp() string {
	log.Println("start ip ===============================")
	// 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Error getting network interfaces:", err)
		return ""
	}

	log.Println("network interfaces len: ", strconv.Itoa(len(interfaces)))

	// 遍历每个网络接口
	for _, iface := range interfaces {
		// 遍历每个地址
		for _, addr := range iface.Addrs {
			log.Println("addr: ", addr.Addr)
		}
	}

	return "127.0.0.1"
}
