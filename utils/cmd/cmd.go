package cmd

import (
	"fmt"
	"ipam/utils/logging"
	"net/netip"
	"os/exec"
	"time"

	"context"

	"github.com/go-ping/ping"

	"go4.org/netipx"
)

func RunShell(cmd string) (result string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//args := strings.Split(cmd, " ")
	cmdCtx := exec.CommandContext(ctx, "bash", "-c", cmd)
	// cmdCtx := exec.CommandContext(ctx, "sshpass", "-p", "VYzB33Lv9g4b", "ssh", "readonly@192.168.169.1", "dis arp vlan 7")
	out, err := cmdCtx.Output()
	if ctx.Err() == context.DeadlineExceeded {
		result = fmt.Sprintf("exec Command timed out")
		return
	}

	return string(out), err
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

