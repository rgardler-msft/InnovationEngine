# Nested Prerequisite Level 2: Should need to run

This nested prerequisite sets the `NESTED_L2_RAN` variable so that its parent
nested prerequisite can validate that this deepest prerequisite has executed.

## Setup

This step sets the `NESTED_L2_RAN` environment variable when the prerequisite needs to run.

```bash
export NESTED_L2_RAN=1
```

## Verification

This verification causes the prerequisite body to execute whenever `NESTED_L2_RAN`
has not been set yet.

```bash
if [ -n "$NESTED_L2_RAN" ]; then
    echo "Nested level 2 prerequisite already executed"
fi

echo "Nested level 2 prerequisite needs to run NESTED_L2_RAN is not set."
```

If the similarity test fails then the main content of this document will be executed.

<!-- expected_similarity=".*already executed" -->
```text
Nested level 2 prerequisite already executed
```
