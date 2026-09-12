package server

import (
	"context"
	"fmt"
	"os/exec"
	"rack-cli/internal/entity"
	"rack-cli/internal/storage"
	"regexp"
	"strconv"
	"time"
)

type Storage = storage.Repo[entity.Server]

var pingRTT = regexp.MustCompile(`time[=<]([0-9.]+)\s*ms`)

func SSHCommand(srv entity.Server) *exec.Cmd {
	return exec.Command("ssh", srv.RootUser+"@"+srv.IP)
}

func Ping(ctx context.Context, srv entity.Server) (time.Duration, error) {
	out, err := exec.CommandContext(ctx, "ping", "-c", "1", srv.IP).CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		return 0, fmt.Errorf("host unreachable")
	}

	m := pingRTT.FindSubmatch(out)
	if m == nil {
		return 0, nil
	}

	ms, _ := strconv.ParseFloat(string(m[1]), 64)
	return time.Duration(ms * float64(time.Millisecond)), nil
}
