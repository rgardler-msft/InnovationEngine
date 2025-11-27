# Execution Environments

Innovation Engine adapts its runtime behavior based on the `--environment` flag (or the `IE_ENVIRONMENT` variable in CI). Most users stick with the default `local`, but hosted scenarios such as GitHub Actions or the Azure portal rely on different modes. This reference explains what each option does and when to use it.

## Quick Reference

| Value            | Description | Typical Usage |
|------------------|-------------|----------------|
| `local`          | Fully interactive terminal experience with minimal cleanup heuristics. | Day-to-day authoring and validation on a workstation. |
| `github-action`  | Non-interactive Bubble Tea renderer, deterministic logging for CI, relaxed terminal assumptions. | GitHub Actions or any CI job that runs `ie test` / `ie execute`. |
| `ocd`            | "One-Click Deployment" mode used by the Azure portal experience. Emits OCD status beacons (`ie_us…ie_ue`), prunes temp state on success, and suppresses interactive prompts. | Portal-backed executions or other services that ingest OCD status lines. |
| `azure`          | Generic Azure-hosted shell/VM. Shares most behavior with `ocd` (non-interactive defaults, remote cleanup) but omits portal-specific telemetry. | Azure Cloud Shell, custom VM agents, or automation where OCD signals are not needed. |

## Local (default)
- Activated automatically when no `--environment` flag is passed.
- Enables interactive Bubble Tea UI, including spinners, keyboard shortcuts, and command previews.
- Leaves `/tmp/ie-env-vars` and `/tmp/working-dir` on disk for inspection after the run.
- Intended for authors iterating on docs or operators running ad hoc executions from their terminals.

## GitHub Action
- Select via `ie <command> doc.md --environment github-action` or set `IE_ENVIRONMENT=github-action` in the workflow.
- Disables alternate screen buffers and other terminal control sequences so logs render cleanly in the Actions viewer.
- Keeps stdout/stderr streaming so `set-output`/log-parsing steps remain deterministic.
- Use this when wiring `ie test` or `ie execute` into CI/CD pipelines.

## OCD (One-Click Deployment)
- The Azure portal wraps Innovation Engine runs in this mode.
- Emits status beacons (`ie_us{json}ie_ue`) parsed by the portal to show progress.
- Automatically collects resource URIs (when available) and prunes temp state files at the end of a run.
- Treat it as "portal automation" mode; you typically never pass this flag manually unless you are building an OCD-compatible surface.

## Azure
- Covers Azure-hosted shells/VMs that are _not_ the portal. Behaves like `ocd` for non-interactive defaults, remote cleanup, and Azure-specific telemetry, but skips OCD beacons.
- Use when you run IE from Azure Cloud Shell, an Azure Container Instance, or any managed VM where ANSI interactivity is limited.

## Selecting an Environment
```bash
ie execute my-doc.md --environment <local|github-action|ocd|azure>
```
If the flag is omitted, `local` is assumed. For automation scenarios, you can also set the `IE_ENVIRONMENT` variable so every CLI invocation shares the same default:

```bash
export IE_ENVIRONMENT=github-action
ie test docs/tutorial.md
```

`ie --help` always lists the available values; consult this reference when deciding which mode matches your runtime.
