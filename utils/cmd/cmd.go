package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"ipam/utils/logging"
	"net/netip"
	"os"
	"os/exec"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/go-ping/ping"

	"go4.org/netipx"
)

func RunShell(cmd string) (result string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	//args := strings.Split(cmd, " ")
	cmdCtx := exec.CommandContext(ctx, "bash", "-c", cmd)
	// cmdCtx := exec.CommandContext(ctx, "sshpass", "-p", "VYzB33Lv9g4b", "ssh", "readonly@192.168.169.1", "dis arp vlan 7")
	out, err := cmdCtx.Output()
	if ctx.Err() == context.DeadlineExceeded {
		result = "exec Command timed out"
		logging.Error("exec Command timed out")
		return
	}

	return string(out), err
}

func RunShellHW(user, password, host string, cmds []string) (string, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		Timeout:         20 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		return "", fmt.Errorf("failed to dial: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = os.Stderr

	err = session.Shell()
	if err != nil {
		return "", fmt.Errorf("failed to start shell: %v", err)
	}

	for _, cmd := range cmds {
		_, err = io.WriteString(stdin, cmd+"\n")
		if err != nil {
			return "", fmt.Errorf("failed to write command: %v", err)
		}
	}
	// 添加退出命令
	_, err = io.WriteString(stdin, "quit\n")
	if err != nil {
		return "", fmt.Errorf("failed to write exit command: %v", err)
	}

	_ = session.Wait()

	return stdoutBuf.String(), nil
}

func PingNetwork(cidr string) error {
	ipnet, err := netip.ParsePrefix(cidr)
	if err != nil {
		return err
	}
	iprange := netipx.RangeOfPrefix(ipnet)
	for ip := iprange.From(); ipnet.Contains(ip); ip = ip.Next() {
		ipstring := ip.String()
		go pingIP(ipstring)
	}
	return nil
}

func pingIP(ip string) {
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		logging.Error("创建 Pinger 错误:", err)
		return
	}
	pinger.SetPrivileged(true)
	pinger.Count = 1
	pinger.Timeout = time.Second
	err = pinger.Run()
	if err != nil {
		logging.Error("执行 Ping 错误:", err)
		return
	}
}
