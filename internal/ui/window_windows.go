//go:build windows

package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func OpenWindow(url, dataDir string) error {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return err
	}
	hta := filepath.Join(dataDir, "agentdesk.hta")
	html := fmt.Sprintf(`<html>
<head>
<title>BiAI AgentDesk</title>
<HTA:APPLICATION ID="AgentDesk" APPLICATIONNAME="BiAI AgentDesk" BORDER="thin" CAPTION="yes" SHOWINTASKBAR="yes" SINGLEINSTANCE="yes" SYSMENU="yes" WINDOWSTATE="normal" />
<script language="javascript">
window.resizeTo(1100, 760);
window.moveTo((screen.width-1100)/2, (screen.height-760)/2);
</script>
</head>
<frameset rows="*">
<frame src="%s" frameborder="0" />
</frameset>
</html>`, htmlEscapeAttr(url))
	if err := os.WriteFile(hta, []byte(html), 0o600); err != nil {
		return err
	}
	cmd := exec.Command("mshta.exe", hta)
	if err := cmd.Start(); err != nil {
		_ = openDefaultBrowser(url)
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case err := <-done:
		_ = openDefaultBrowser(url)
		if err != nil {
			return fmt.Errorf("mshta exited immediately: %w", err)
		}
		return fmt.Errorf("mshta exited immediately")
	case <-time.After(1500 * time.Millisecond):
		return nil
	}
}

func htmlEscapeAttr(s string) string {
	r := strings.NewReplacer("&", "&amp;", `"`, "&quot;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

func openDefaultBrowser(url string) error {
	return exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", url).Start()
}
