package main

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type TcpLogger struct {
	conn    *net.Conn
	mutex   sync.Mutex
	Address string
}


func (l TcpLogger) Sync() error {
	return nil
}

func (l TcpLogger) Write(p []byte) (int, error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.conn == nil {
		conn, err := net.Dial("tcp", l.Address)
		if err != nil {
			return 0, err
		}

		l.conn = &conn
	}
	contentLength := len(p)
	written := 0
	for written < contentLength {
		n, err := (*l.conn).Write(p[written:])
		if err != nil {
			return written, err
		}
		written += n
	}

	return written, nil
}



func main() {
	writeFile := zapcore.AddSync(&lumberjack.Logger{
		Filename: "service.log",
		MaxSize:  10, // megabytes
		MaxAge:   28, // days
	})
	writeStdout := zapcore.AddSync(os.Stdout)
	writeTcp := zapcore.AddSync(&TcpLogger{
		Address: "127.0.0.1:12345",
	})

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.NewMultiWriteSyncer(writeFile, writeStdout, writeTcp),
		zap.InfoLevel,
	)

	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)
	defer logger.Sync()

	sigs := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <- sigs
		logger.Info("Receive signal: ", zap.String("signal", sig.String()))
		done <- struct{}{}
	}()
	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context) {
		for {
			select {
			case <- ctx.Done():
				return
			default:
				logger.Info("I am a test message")
				time.Sleep(time.Second * 2)
			}
		}
	}(ctx)

	logger.Info("Waiting to receive signals")
	<- done
	cancel()
	logger.Info("Done")
}