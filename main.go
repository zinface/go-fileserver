package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
)

// ---------------------------------------------------- 基本数据结构定义

const (
	// DATAPART 数据分片大小
	DATAPART = 4096
	
)
// Package 数据包内容定义
type Package struct {
	datalen uint32 
	databuf []byte
}

// FileInfo 文件基础信息
type FileInfo struct {
	filename   string
	filelength int32
	readStart  int64

	fp     *os.File
	buffer []byte
}

// SocketInfo 套接字基础信息
type SocketInfo struct {
	ip      string
	port    int
	timeout int
}

// ---------------------------------------------------- 数据包内容创建
var simplePackage = &Package{datalen: 0, databuf: nil}

func (p *Package) buildSimplePackge() {
	p.datalen = uint32(len(currentFile.filename))
	p.databuf = []byte(currentFile.filename)
}

func (p *Package) sendSimplePackge(conn net.Conn) {
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.LittleEndian, p.datalen+1)
	buffer.Write(p.databuf)
	conn.Write(buffer.Bytes())
	buffer.Reset()

	var buf = make([]byte, 4)
	conn.Read(buf)
	buffer.Write(buf)

	var _start uint32
	binary.Read(&buffer, binary.LittleEndian, &_start)
	if int32(_start) != -1 {
		currentFile.readStart = int64(_start)
		currentFile.filelength = currentFile.filelength - int32(_start)
	} else {
		currentFile.readStart = 0
	}
	buffer.Reset()

	binary.Write(&buffer, binary.LittleEndian, currentFile.filelength)
	conn.Write(buffer.Bytes())
}

// ---------------------------------------------------- net/ socket 构建
var currentSocket = &SocketInfo{}
var currentListener = &net.TCPListener{}

func (cs *SocketInfo) getHost() string {
	if cs.ip == "" {
		return fmt.Sprintf(":%d", cs.port)
	}
	return fmt.Sprintf("%s:%d", cs.ip, cs.port)
}

func (cs *SocketInfo) newSocketListener() *net.TCPListener {
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", cs.getHost())
	listener, _ := net.ListenTCP("tcp", tcpAddr)
	return listener
}

// ---------------------------------------------------- file/ 读写 构建

var currentFile = &FileInfo{}

func (f *FileInfo) doExchange() {
	conn, _ := currentListener.Accept()


// send 基础信息
fmt.Println("// send 基础信息")
	simplePackage.buildSimplePackge()
	simplePackage.sendSimplePackge(conn)

	// 切片分段整(4096)
	var integerSize = currentFile.filelength / DATAPART
	// 切片分段余(4096)
	var remainderSize = currentFile.filelength % DATAPART

	fmt.Printf("    // send 切片分段整: %d 片(%d字节/片)\n", integerSize, DATAPART)
	fmt.Printf("    // send 切片分段余: %d 字节\n", remainderSize)

// send 分段整
fmt.Print("// send 分段整\r"); os.Stdout.Sync()
	var cnt = currentFile.readStart
	for {
		if cnt == int64(integerSize)*DATAPART + currentFile.readStart {
			break
		}
		f.read(cnt)
		conn.Write(f.buffer)
		cnt += DATAPART
		fmt.Printf("// send 分段整: %d / %d\r", integerSize, cnt/DATAPART)
		os.Stdout.Sync()
	}
	fmt.Println("")
// send 分段余
fmt.Print("// send 分段余\r"); os.Stdout.Sync()
	f.read(cnt)
	conn.Write(f.buffer)
	fmt.Printf("// send 分段余: %d / %d\r", remainderSize, remainderSize)
	os.Stdout.Sync()
}

func (f *FileInfo) read(size int64) {
	f.fp.Seek(size, os.SEEK_SET)
	f.fp.Read(f.buffer)
}

func (f *FileInfo) write() {

}

func (f *FileInfo) loadFileLength() (int32, error) {
	fp, _ := os.Open(currentFile.filename)
	f.fp = fp
	size, err :=  f.fp.Seek(0, os.SEEK_END)
	return int32(size), err
}

func (f *FileInfo) loadFile(filename string) {
	currentFile.filename = filename

	if _, err := os.Stat(currentFile.filename); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "os.Stat error: '%s' 文件不存在\n", currentFile.filename)
		os.Exit(1)
	}
	if fs, _ := os.Stat(currentFile.filename); fs.IsDir() {
		fmt.Fprintf(os.Stderr, "os.Stat error: '%s' 是一个目录\n", currentFile.filename)
		os.Exit(1)
	}

	currentFile.filelength, _ = currentFile.loadFileLength()
}

// ---------------------------------------------------- 主程序入口/及初始化动作

func init() {
	currentSocket.ip = "0.0.0.0"
	currentSocket.port = 8082
	currentSocket.timeout = 3000

	currentFile.filename = "tmp.txt"
	currentFile.filelength = 999
	currentFile.readStart = 0
	currentFile.buffer = make([]byte, DATAPART)

	currentListener = currentSocket.newSocketListener()
}

func usage(args []string) {
	if len(args) < 2 {
		fmt.Printf("Usage: %s [FILE]...\n\n", args[0])
		os.Exit(1)
	}
}

func main() {
	usage(os.Args)

	currentFile.loadFile(os.Args[1])

	currentFile.doExchange()
}
