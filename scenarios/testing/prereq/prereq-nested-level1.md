# Nested Prerequisite Level 1: Chains to level 2

This prerequisite document is used to test multi-level nesting of prerequisites.
It has its own setup and verification, and it also declares another prerequisite
(`prereq-nested-level2.md`) in its prerequisites section.

## Prerequisites

This nested prerequisite depends on another prerequisite that should be executed first.

 [Nested prerequisite level 2](prereq-nested-level2.md)

## Setup

When this prerequisite needs to run, it sets the `NESTED_L1_RAN` environment variable.

```bash
export NESTED_L1_RAN=1
```

## Verification

This verification causes the prerequisite body to execute whenever `NESTED_L1_RAN`
has not been set yet.

```bash
if [ -n "$NESTED_L1_RAN" ]; then
    echo "Nested level 1 prerequisite already executed"
fi

echo "Nested level 1 prerequisite needs to run NESTED_L1_RAN is not set."
```

If the similarity test fails then the main content of this document will be executed.

<!-- expected_similarity=".*already executed" -->
```text
Nested level 1 prerequisite already executed
```
