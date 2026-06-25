package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type dashboardOpenBody struct {
	// Seconds before the backend kills the kiosk browser and systemd restarts it
	// back to /kiosk. Zero means no auto-return; the user must navigate manually.
	TimeoutSecs int `json:"timeout_secs" minimum:"0" doc:"Seconds before auto-returning to the frame; 0 disables auto-return"`
}

type dashboardOpenInput struct {
	Body dashboardOpenBody
}

func (s *server) registerDashboardRoutes(api huma.API) {
	s.kioskExempt("/api/kiosk/dashboard")
	huma.Register(api, huma.Operation{
		OperationID:   "open-dashboard",
		Method:        http.MethodPost,
		Path:          "/api/kiosk/dashboard",
		Summary:       "Record dashboard open and schedule auto-return",
		DefaultStatus: http.StatusNoContent,
	}, func(_ context.Context, input *dashboardOpenInput) (*struct{}, error) {
		s.dashboardMu.Lock()
		defer s.dashboardMu.Unlock()

		// Cancel any timer left from a previous open.
		if s.dashboardCancel != nil {
			s.dashboardCancel()
			s.dashboardCancel = nil
		}

		if input.Body.TimeoutSecs <= 0 {
			return nil, nil
		}

		ctx, cancel := context.WithCancel(context.Background())
		s.dashboardCancel = cancel

		go func() {
			select {
			case <-ctx.Done():
				// Cancelled: kiosk resumed before the timeout (heartbeat fired).
				return
			case <-time.After(time.Duration(input.Body.TimeoutSecs) * time.Second):
				// Timeout elapsed: kill cog so systemd restarts it at /kiosk.
				returnBrowserToKiosk(s.log)
			}
		}()

		return nil, nil
	})
}

// cancelDashboardReturn stops any pending auto-return timer.
// Called when a heartbeat arrives from the on-device kiosk, which means the
// user has navigated back to /kiosk on their own before the timeout fired.
func (s *server) cancelDashboardReturn() {
	s.dashboardMu.Lock()
	defer s.dashboardMu.Unlock()
	if s.dashboardCancel != nil {
		s.dashboardCancel()
		s.dashboardCancel = nil
	}
}

// returnBrowserToKiosk terminates the cog browser process.
// Both kiosk-backend and kiosk-browser run as the same OS user, so no elevated
// permissions are required. Systemd restarts cog (RestartSec=10) and it loads
// http://localhost/kiosk — returning the user to the picture frame.
func returnBrowserToKiosk(log interface{ Info(string, ...any); Warn(string, ...any) }) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		log.Warn("dashboard return: cannot read /proc", "err", err)
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		comm, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(comm)) == "cog" {
			if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
				log.Warn("dashboard return: failed to signal cog", "pid", pid, "err", err)
			} else {
				log.Info("dashboard return: sent SIGTERM to cog, systemd will restart it", "pid", pid)
			}
			return
		}
	}
	log.Warn("dashboard return: cog process not found in /proc")
}
