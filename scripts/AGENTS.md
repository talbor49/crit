# Scripts

## e2e-roundtrip.sh

These are the provider-specific runners for one live roundtrip suite in `internal/session`. GitHub uses the `e2e_github` tag, `CRIT_ROUNDTRIP_REPO`, and authenticated `gh`; GitLab uses the `e2e_gitlab` tag, `CRIT_GITLAB_ROUNDTRIP_PROJECT`, and authenticated `glab` (plus optional `CRIT_GITLAB_ROUNDTRIP_HOST` for self-managed instances). Both create and clean up a temporary change request and branch. See `test/roundtrip/README.md` for shared setup and authoring guidance.
