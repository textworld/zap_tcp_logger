package main

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"net"
	"os"
	"sync"
)

type TcpLogger struct {
	conn *net.Conn
	mutex sync.Mutex
}

func CreateTcpLogger(address string) (TcpLogger){
	tcpLogger := TcpLogger{}
	conn, err := net.Dial("tcp", address)

	tcpLogger.conn conn
	return tcpLogger
}

func (l TcpLogger) Sync() error {
	return nil
}

func (l TcpLogger) Write(p []byte) (int, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.conn == nil {

	}
}



func main() {
	writeFile := zapcore.AddSync(&lumberjack.Logger{
		Filename: "service.log",
		MaxSize:  10, // megabytes
		MaxAge:   28, // days
	})
	writeStdout := zapcore.AddSync(os.Stdout)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.NewMultiWriteSyncer(writeFile, writeStdout),
		zap.InfoLevel,
	)

	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)
	defer logger.Sync()

	logger.Info("Test")
}