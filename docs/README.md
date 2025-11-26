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
  2. 

  3. [Build a Hello World script](tutorial/README.md)

