# Executable Documentation Quickstart

## Introduction
Executable Documentation Quickstart guides you through cloning the Innovation Engine repository, building the CLI, and running a smoke test against a real scenario so you can see verified output immediately. Treat this file as a living template—swap in your own prerequisites, commands, and validation steps while keeping the proven structure intact.

Completing this walkthrough leaves you with a timestamped local workspace, a compiled `ie` binary, and a reproducible log you can share in reviews or attach to CI artifacts.

## Prerequisites
You need a Unix-like shell with Git, Make, and Go 1.21+ available so the CLI can be built from source. Docker is optional unless you plan to build container images as part of your scenario.

Before you begin, confirm that you can run `git --version`, `make --version`, and `go version` without errors. If any of these commands fail, install the missing dependency and then return to this document so your walkthrough starts from a known-good baseline.

## Setting up the environment
Every command in this document relies on explicit environment variables for clarity and repeatability. Adjust the defaults as needed, but keep the `_HASH` suffix so reruns do not collide with prior artifacts.

| Variable | Default | Description |
| --- | --- | --- |
| `HASH` | `2511251200` | Timestamp (YYMMDDHHMM) used to uniquify directories and log files; regenerate per run with `date +"%y%m%d%H%M"`. |
| `IE_REPO_URL` | `https://github.com/Azure/InnovationEngine.git` | Source repository that hosts the `ie` CLI and example docs. |
| `IE_CLONE_DIR` | `$HOME/innovation-engine_${HASH}` | Workspace destination; suffix avoids clobbering previous clones. |
| `IE_BIN_PATH` | `$IE_CLONE_DIR/bin/ie` | Location of the compiled CLI once `make build-ie` finishes. |
| `IE_SCENARIO_PATH` | `$IE_CLONE_DIR/scenarios/testing/test.md` | Smoke-test document that exercises core engine features. |
| `IE_MODE` | `execute` | Default execution mode for the walkthrough; switch to `interactive` or `test` as needed. |
| `IE_LOG_LEVEL` | `info` | Structured logging threshold for the run. |
| `IE_LOG_PATH` | `$IE_CLONE_DIR/logs/ie.log` | Output file for persisted logs; directories are created automatically. |

Export these variables into your shell session before moving on so every subsequent code block can rely on them without inline edits, keeping the remainder of the document copy/paste friendly.

## Steps
This section walks chronologically through cloning, building, executing, and cleaning up. Run the blocks in order; each summary highlights what changed and what to verify before moving forward.

### Step 1: Clone the repository
Clone the Innovation Engine repository into a timestamped directory so you have an isolated workspace for experimentation.

```bash
git clone $IE_REPO_URL $IE_CLONE_DIR
```

You should now have a dedicated `$IE_CLONE_DIR` populated with the CLI source, executable documentation examples, and supporting assets.

### Step 2: Build the CLI
Compile the `ie` binary from source. The build also installs dependencies declared in `go.mod` so the resulting executable is self-contained.

```bash
cd $IE_CLONE_DIR;
make build-ie;
```

After the build succeeds you will find the CLI at `$IE_BIN_PATH`; confirm it exists before proceeding.

### Step 3: Run the smoke test
Execute the master testing scenario to validate code block handling, prerequisite validation, and result comparison. Logs are persisted for later inspection.

```bash
mkdir -p $(dirname $IE_LOG_PATH);
$IE_BIN_PATH $IE_MODE $IE_SCENARIO_PATH --log-level $IE_LOG_LEVEL --log-path $IE_LOG_PATH;
```

Review the console output (and `$IE_LOG_PATH`) to ensure the scenario reported success; failures typically point to missing tools or environment configuration.

### Step 4: (Optional) Clean up
Remove the working directory if you want to reclaim space once the walkthrough is complete.

```bash
rm -rf $IE_CLONE_DIR
```

Only run this step after you have captured any artifacts you need, such as logs or modified documents.

After completing each step you now have a repeatable workflow to bootstrap new executable documentation scenarios, complete with deterministic paths and log locations.

## Summary
You cloned Innovation Engine into a unique directory, built the CLI from source, ran a comprehensive smoke test, and optionally cleaned up the workspace. The environment variables you configured ensure those actions remain reproducible across machines and automation.

## Next Steps
Extend this template by swapping in your own scenario paths, prerequisites, or validation logic. Useful follow-ups include:

1. Author a new tutorial starting from `docs/helloWorldDemo.md` and run it through `ie test`.
2. Wire the smoke test into CI by invoking the same commands in GitHub Actions or Azure Pipelines.
3. Explore `docs/prerequisitesAndIncludes.md` to incorporate reusable prerequisite documents.

Adapting the template in these ways helps you scale executable documentation across teams without re-learning the wiring each time.
