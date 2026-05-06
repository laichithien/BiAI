package system

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
)

func DynamicContext(dataDir, workspace string) string {
	host, _ := os.Hostname()
	wd, _ := os.Getwd()
	u, _ := user.Current()
	username := ""
	if u != nil {
		username = u.Username
	}
	return fmt.Sprintf(`Runtime context:
- OS: %s
- Arch: %s
- Go runtime: %s
- Hostname: %s
- User: %s
- App data dir: %s
- Process working dir: %s
- Selected workspace: %s
- Network/API access depends on configured LLM endpoint.
- Filesystem tools are restricted to selected workspace.
- Destructive commands require app approval before execution.`,
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
		host,
		username,
		dataDir,
		wd,
		workspace,
	)
}
