package gitlab

import (
	"fmt"

	"github.com/tomasz-tomczyk/crit/internal/session"
)

// checkGitLabSyncAllowed gates `crit pull` and `crit push` from running on
// live reviews. Live pins have no line anchors and cannot round-trip through
// GitLab MR discussion notes.
func checkGitLabSyncAllowed(cj session.CritJSON, op string) error {
	if cj.ReviewType == "live" {
		return fmt.Errorf("%s is not supported for live reviews", op)
	}
	return nil
}
