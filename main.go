package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		parent()
	case "child":
		child()
	default:
		panic("what should i do")
	}
}

func parent() {
	// /proc/n/exe はプロセスの実行ファイルへのシンボリックリンク
	// /proc/self はプロセス自身を指す
	// つまり自分自身を(第1引数に `child` を足しつつ)再実行する
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	// プログラムを UTS/PID/MNT Namespace 内で実行する
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
}

func child() {
	must(syscall.Mount("rootfs", "rootfs", "", syscall.MS_BIND, ""))
	must(os.MkdirAll("rootfs/oldrootfs", 0700))
	must(syscall.PivotRoot("rootfs", "rootfs/oldrootfs"))
	must(os.Chdir("/"))
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
