package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/go-ps"
	"golang.org/x/sys/unix"
)

// main 函数是程序的入口点，负责初始化和启动配置监视器。
func main() {
	// 从环境变量中获取配置目录
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		log.Fatal("必填环境变量 CONFIG_DIR 为空，程序退出")
	}

	// 从环境变量中获取进程名称
	processName := os.Getenv("PROCESS_NAME")
	if processName == "" {
		log.Fatal("必填环境变量 PROCESS_NAME 为空，程序退出")
	}

	// 是否启用详细模式的标志
	verbose := false
	verboseFlag := os.Getenv("VERBOSE")
	if verboseFlag == "true" {
		verbose = true
	}

	// 获取重新加载信号，默认为 SIGHUP
	var reloadSignal syscall.Signal
	reloadSignalStr := os.Getenv("RELOAD_SIGNAL")
	if reloadSignalStr == "" {
		log.Printf("RELOAD_SIGNAL 为空，将默认为 SIGHUP")
		reloadSignal = syscall.SIGHUP
	} else {
		reloadSignal = unix.SignalNum(reloadSignalStr)
		if reloadSignal == 0 {
			log.Fatalf("无法找到 RELOAD_SIGNAL 的信号：%s", reloadSignalStr)
		}
	}

	// 打印启动信息
	log.Printf("启动，CONFIG_DIR=%s, PROCESS_NAME=%s, RELOAD_SIGNAL=%s\n", configDir, processName, reloadSignal)

	// 创建文件系统监视器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// 用于在监视结束时发出信号的通道
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if verbose {
					log.Println("事件:", event)
				}
				if event.Op&fsnotify.Chmod != fsnotify.Chmod {
					log.Println("修改的文件:", event.Name)
					err := reloadProcess(processName, reloadSignal)
					if err != nil {
						log.Println("错误:", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("错误:", err)
			}
		}
	}()

	// 将要监视的目录添加到监视器
	configDirs := strings.Split(configDir, ",")
	for _, dir := range configDirs {
		err = watcher.Add(dir)
		if err != nil {
			log.Fatal(err)
		}
	}

	// 等待监视结束
	<-done
}

// findPID 查找给定进程名称的进程 ID（PID）。
func findPID(process string) (int, error) {
	processes, err := ps.Processes()
	if err != nil {
		return -1, fmt.Errorf("无法列出进程：%v\n", err)
	}

	for _, p := range processes {
		if p.Executable() == process {
			log.Printf("找到可执行文件 %s（pid：%d）\n", p.Executable(), p.Pid())
			return p.Pid(), nil
		}
	}

	return -1, fmt.Errorf("未找到匹配 %s 的进程\n", process)
}

// reloadProcess 使用给定的信号重新加载指定的进程。
func reloadProcess(process string, signal syscall.Signal) error {
	pid, err := findPID(process)
	if err != nil {
		return err
	}

	err = syscall.Kill(pid, signal)
	if err != nil {
		return fmt.Errorf("无法发送信号：%v\n", err)
	}

	log.Printf("信号 %s 已发送至 %s（pid：%d）\n", signal, process, pid)
	return nil
}
