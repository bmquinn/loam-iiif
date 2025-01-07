package iiif

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/bmquinn/loam-iiif/internal/types"
	tea "github.com/charmbracelet/bubbletea"
)

func FetchData(urlStr string) tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get(urlStr)
		if err != nil {
			return types.ErrMsg{Error: err}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return types.ErrMsg{Error: fmt.Errorf("failed to fetch data: %s", resp.Status)}
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return types.ErrMsg{Error: err}
		}
		return types.FetchDataMsg(body)
	}
}

func OpenURL(urlStr string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{urlStr}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", urlStr}
	default:
		cmd = "xdg-open"
		args = []string{urlStr}
	}

	return exec.Command(cmd, args...).Start()
}
