package main

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reddit/dao/mysql"
	"reddit/dao/redis"
	"reddit/logger"
	"reddit/routes"
	"reddit/settings"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// go web项目通用脚手架
func main() {
	//1.初始化配置
	if err := settings.Init(); err != nil {
		fmt.Printf("ettings.Init() is failed err:%v\n", err)
	}
	//2.初始化日志文件
	if err := logger.Init(); err != nil {
		fmt.Printf("logger.Init() is failed err:%v\n", err)
	}
	zap.L().Sync()
	zap.L().Debug("logger init success...")
	//3.初始化mysql
	if err := mysql.Init(); err != nil {
		fmt.Printf("mysql.Init() is failed err:%v\n", err)
	}
	defer mysql.Close()

	//4.初始化redis
	if err := redis.Init(); err != nil {
		fmt.Printf("redis.Init() is failed err:%v\n", err)
	}
	defer redis.Close()
	//5.初始化路由
	//6.启动路由
	r := routes.Signup()
	//7.优雅关机
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.GetInt("app.port")),
		Handler: r,
	}

	go func() {
		// 开启一个goroutine启动服务
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中断信号来优雅地关闭服务器，为关闭服务器操作设置一个5秒的超时
	quit := make(chan os.Signal, 1) // 创建一个接收信号的通道
	// kill 默认会发送 syscall.SIGTERM 信号
	// kill -2 发送 syscall.SIGINT 信号，我们常用的Ctrl+C就是触发系统SIGINT信号
	// kill -9 发送 syscall.SIGKILL 信号，但是不能被捕获，所以不需要添加它
	// signal.Notify把收到的 syscall.SIGINT或syscall.SIGTERM 信号转发给quit
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 此处不会阻塞
	<-quit                                               // 阻塞在此，当接收到上述两种信号时才会往下执行
	zap.L().Info("Shutdown Server ...")
	// 创建一个5秒超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// 5秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过5秒就超时退出
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Error("Server Shutdown: ", zap.Error(err))
	}
	zap.L().Info("Server exiting")
}
