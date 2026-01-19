# mbserver

[![Go Version](https://img.shields.io/github/go-mod/go-version/leijux/mbserver)](https://github.com/leijux/mbserver)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/leijux/mbserver.svg)](https://pkg.go.dev/github.com/leijux/mbserver)

一个用 Go 语言实现的 Modbus 服务器（从站），支持 TCP 和 RTU（串行）协议。

## 功能特性

- 完整的 Modbus 协议支持（功能码 1、2、3、4、5、6、15、16）
- 支持 TCP、TLS 和 RTU（串行）传输层
- 可自定义的内存寄存器（线圈、离散输入、保持寄存器、输入寄存器）
- 可扩展的函数处理器（支持自定义功能码）
- 线程安全，并发处理请求
- 优雅关闭

## 安装

```bash
go get github.com/leijux/mbserver
```

## 快速开始

以下是一个简单的 TCP 服务器示例：

```go
package main

import (
    "context"
    "flag"
    "log/slog"
    "os/signal"
    "syscall"

    "github.com/leijux/mbserver"
)

var addr = flag.String("addr", ":8080", "TCP address to listen on")

func main() {
    flag.Parse()

    ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

    s := mbserver.NewServer()

    err := s.ListenTCP(*addr)
    if err != nil {
        slog.Error("listen tcp err", "err", err)
        return
    }

    defer s.Shutdown()

    go s.Start()

    <-ctx.Done()
}
```

运行该程序将启动一个监听 `:8080` 的 Modbus TCP 服务器。

## 使用方法

### 创建服务器

```go
s := mbserver.NewServer()
```

### 使用自定义寄存器

默认情况下，服务器使用内存寄存器（每个区域 65536 个地址）。你可以提供自己的寄存器实现：

```go
type MyRegister struct {
    // 实现 mbserver.Register 接口
}

mr := &MyRegister{}
s := mbserver.NewServer(mbserver.WithRegister(mr))
```

### 自定义函数处理器

你可以为特定的功能码注册自定义处理函数：

```go
s := mbserver.NewServer(mbserver.WithRegisterFunction(0x41, myCustomFunction))
```

### 监听 TCP

```go
err := s.ListenTCP(":502")
if err != nil {
    // 处理错误
}
```

### 监听串行端口（RTU）

```go
import "github.com/goburrow/serial"

config := &serial.Config{
    Address:  "/dev/ttyUSB0",
    BaudRate: 9600,
    DataBits: 8,
    StopBits: 1,
    Parity:   "N",
}
err := s.ListenRTU(config)
if err != nil {
    // 处理错误
}
```

### 监听 TLS（安全 TCP）

```go
import "crypto/tls"

tlsConfig := &tls.Config{
    // 配置 TLS 证书和密钥
}
err := s.ListenTLS(":802", tlsConfig)
if err != nil {
    // 处理错误
}
```

### 启动服务器

```go
go s.Start()
```

### 关闭服务器

```go
s.Shutdown()
```

## API 文档

完整的 API 文档请参阅 [pkg.go.dev/github.com/leijux/mbserver](https://pkg.go.dev/github.com/leijux/mbserver)。

## 示例

更多示例请查看 `cmd/` 目录和测试文件。

## 贡献

欢迎提交 Issue 和 Pull Request。

## 许可证

本项目基于 MIT 许可证开源，详见 [LICENSE](LICENSE) 文件。