echo $"Hello World"
# Documentation Guide

This directory aggregates everything you need to learn, author, and test executable documentation with Innovation Engine. Use it as a curated map rather than searching the tree by hand.

## Start Here
Begin with the foundational walkthroughs if you are new to Innovation Engine:

- `../README.md` – platform overview, install options, and CLI basics.
- `helloWorldDemo.md` – minimal example that shows command execution paired with expected output blocks.
- `Executable-Doc-Quickstart.md` – opinionated template that scaffolds prerequisites, environment variables, and deterministic steps.

These three files get you from zero to a validated executable document while highlighting the authoring patterns the rest of the docs assume.

## Authoring Playbooks
Use these guides when you are ready to design richer scenarios or factor out reusable building blocks:

- `Authoring-With-Copilot.md` – workflow for pairing the IE CLI with GitHub Copilot during content creation.
- `prerequisitesAndIncludes.md` plus `Common/prerequisiteExample.md` – how to structure prerequisite documents, verification sections, and includes.
- `Common/environmentVariablesFromPrerequisites.md` – patterns for sharing environment configuration across documents safely.

Following the playbooks keeps large tutorials maintainable and repeatable across teams.

## Reference
Keep these files nearby when you need specific API or runtime details:

- `environmentVariables.md` and `environmentVariables.ini` – catalog of built-in variables and sample defaults.
- `environments.md` – explains the `--environment` flag (local, GitHub Action, OCD, Azure) and when to select each mode.
- `modesOfOperation.md` – deeper dive into `execute`, `interactive`, and `test` behaviors (pausing rules, failure semantics, etc.).

They pair well with the codebase when you are debugging or extending the engine itself.

## Specs and Testing
Trace expected behaviors and validation strategies through the following assets:

- `specs/test-reporting.md` – requirements for result blocks, similarity scoring, and reporting formats.
- `../scenarios/testing/test.md` – canonical scenario that exercises streaming output, prerequisites, and fuzzy-matching checks.
- `../tests/cli_integration_test.go` – Go-based integration coverage for the CLI entry points.

Treat these as executable acceptance criteria when contributing new engine features.

## Additional Examples
Looking for domain-specific samples? Explore the curated scenario directories under `../examples/` (AKS, authentication, VM operations, and more). Each folder pairs narrative Markdown with runnable commands tailored to that workload.

If you build a new scenario, add it alongside the closest domain example and reference it from this guide so future authors can discover it quickly.

## Example Author → Production Flow
The sequence below shows how executable docs typically move from a draft to running in production environments. Use the IE CLI where noted and your normal repo tooling for the rest.

1. **Author** – Write or update the Markdown scenario and run `ie inspect my-doc.md` to confirm the structure, prerequisite annotations, and descriptions look right before executing anything destructive. `inspect` linting enforces authoring hygiene: it validates language tags, prerequisite `expected_results` blocks (while allowing export-only prereqs), environment variable naming, and usage. You will see warnings for unused exports, and errors if uppercase variables are referenced without a corresponding export or assignment. Lowercase locals (e.g., loop counters) and helper assignments like `VAR=$(command)` are automatically exempt, so keep locals lowercase and export anything intended for reuse. See [Authoring-With-Copilot.md](Authoring-With-Copilot.md) for deeper guidance on structuring content and leveraging assistants during this phase.
2. **Test** – Validate all `expected_results` blocks locally with `ie test my-doc.md`. This executes each code block in a sandbox, comparing actual output to the embedded expectations so you can tighten similarity thresholds early; refer to [specs/test-reporting.md](specs/test-reporting.md) for expectations, similarity scoring, and troubleshooting tips.
3. **Publish** – Commit the validated doc alongside any prerequisite files and open a PR. Augment your repo by wiring a GitHub Action (or equivalent CI job) that runs `ie test` against the PR to block regressions before review completes—see the CI recommendations in [Executable-Doc-Quickstart.md#next-steps](Executable-Doc-Quickstart.md#next-steps).
4. **Execute** – After merge, run the scenario with `ie execute my-doc.md` to provision resources in the target subscription or cluster. Provide per-run values with repeated `--var NAME=value` flags whenever the scenario exposes inputs; [modesOfOperation.md](modesOfOperation.md) covers execute semantics versus other modes.
5. **Env-capture** – Export the resulting configuration into a reusable script: `ie env-config > ie-env.sh`. Consumers can `source ie-env.sh` to hydrate their shells, and you can optionally add `--prefix` if you only want a subset of variables—see [environmentVariables.md](environmentVariables.md) for naming patterns and additional tooling.
6. **Test (post-deploy)** – Use your team’s established QA/validation processes to exercise the freshly provisioned environment (integration tests, smoke tests, monitoring checkpoints, etc.). The goal is to verify the workload itself, so lean on your existing QA runbooks even though there is no IE-specific document for this stage yet.
7. **Export script** – Create a fully standalone shell script with `ie to-bash my-doc.md > deploy.sh` so automation systems without the IE runtime can execute the scenario; consult [modesOfOperation.md#to-bash-mode](modesOfOperation.md#to-bash-mode) for details. In most cases you will pair this script with the captured `ie-env.sh` file so both the actions and their parameters travel together.
8. **Production** – Feed `deploy.sh` (plus the env file) into your production release pipeline or change-management tooling, following the CI/CD integration advice in [Executable-Doc-Quickstart.md#next-steps](Executable-Doc-Quickstart.md#next-steps). The pipeline becomes responsible for scheduling, approvals, and auditing while the generated artifacts keep the actual deployment steps deterministic.
  2. 

  3. [Build a Hello World script](tutorial/README.md)

