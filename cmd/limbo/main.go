//go:build linux

package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	InitUserNamespace = "initUserNamespace"
)

func init() {
	if os.Args[0] == InitUserNamespace {
		mountDir := os.Args[1]
		commands := os.Args[2]
		var stdinReader = bufio.NewReader(os.Stdin)

		log.Printf("[%s] running init user namespace with %s", mountDir, os.Args)
		if err := nsSetup(mountDir, stdinReader); err != nil {
			log.Printf("[%s] error setting up user namespace: %s\n", mountDir, err)
			os.Exit(1)
		}

		if err := nsRun(mountDir, commands); err != nil {
			log.Printf("[%s] error running nested user namespace: %s\n", mountDir, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
}

func nsTeardown(mountDir string) {
	log.Printf("[%s] unmounting %q...", mountDir, mountDir)
	if err := syscall.Unmount(mountDir, 0); err != nil {
		log.Printf("[%s] error unmount %q: %s", mountDir, mountDir, err)
		os.Exit(1)
	}
}

func nsSetup(mountDir string, stdinReader io.Reader) error {
	if err := syscall.Mount("none", mountDir, "tmpfs", syscall.MS_NOSUID, "size=32m"); err != nil {
		return fmt.Errorf("[%s] error mounting tmpfs: %s", mountDir, err)
	}
	f, err := os.OpenFile(filepath.Join(mountDir, "input.tar.gz"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("[%s] error opening file: %s", mountDir, err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	_, err = io.Copy(w, stdinReader)
	if err != nil {
		return fmt.Errorf("[%s] error copying stdin to input.tar.gz file: %s", mountDir, err)
	}
	return nil
}

func nsRun(mountDir, commands string) error {
	defer nsTeardown(mountDir)
	cmds := strings.Split(commands, ";")
	for _, command := range cmds {
		cmdParts := strings.Split(command, " ")
		cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
		cmd.Dir = mountDir
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWUSER,
			UidMappings: []syscall.SysProcIDMap{
				{
					ContainerID: 999,
					HostID:      os.Getuid(),
					Size:        1,
				},
			},
			GidMappings: []syscall.SysProcIDMap{
				{
					ContainerID: 999,
					HostID:      os.Getgid(),
					Size:        1,
				},
			},
		}
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("[%s] error running the command: %s: %s: %s", mountDir, cmdParts, err, output)
		}
		log.Printf("[%s] command %q output: %s", mountDir, command, output)
	}
	return nil
}

func main() {
	log.Println("Starting limbo...")
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("error getting kubernetes config: %s\n", err)
		os.Exit(1)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("error getting kubernetes client: %s\n", err)
		os.Exit(1)
	}

	for {
		Run(client)
		time.Sleep(5000 * time.Millisecond)
	}
}

func Run(clientset *kubernetes.Clientset) {
	cm, err := clientset.CoreV1().ConfigMaps("default").Get(context.Background(), "limbo-cmds", metav1.GetOptions{})
	if err != nil {
		log.Printf("error getting configmap: %s", err)
		return
	}

	f, err := os.OpenFile("busybox.tar.gz", os.O_RDONLY, 0600)
	if err != nil {
		log.Printf("error opening busybox.tar.gz: %s", err)
		return
	}
	for key, entry := range cm.Data {

		cmddir := filepath.Join("/tmp", key)
		os.Mkdir(cmddir, 0755)
		args := append([]string{InitUserNamespace}, cmddir, entry)
		cmd := &exec.Cmd{
			Path: "/proc/self/exe",
			Args: args,
		}
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER,
			UidMappings: []syscall.SysProcIDMap{
				{
					ContainerID: 0,
					HostID:      os.Getuid(),
					Size:        1,
				},
			},
			GidMappings: []syscall.SysProcIDMap{
				{
					ContainerID: 0,
					HostID:      os.Getgid(),
					Size:        1,
				},
			},
		}

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if key == "cmd1" {
			cmd.Stdin = bufio.NewReader(f)
		}

		err = cmd.Run()
		if err != nil {
			log.Printf("error running %q: %s\n", args, err)
		}
	}
}
