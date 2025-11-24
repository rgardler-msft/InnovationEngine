# Test Prerequisites

This document has both present and missing prerequisites. The one that is present should be executed, allowing this document to fully execute. The missing one should only produce a warning and not abort execution.

## Prerequisites

The following prerequisite documents will be executed before validation. One exists, one does not to exercise warning + continue behavior, one includes verification logic that should cause it to skip execution, and one exercises multi-level nested prerequisites.

- [A prerequisite that needs to run](prereq/prereq-set-variable.md)
- [A prereuisite that does not need to run](prereq/prereq-skip-with-verification.md)
- [A nested prerequisite chain](prereq/prereq-nested-level1.md)
- [A prerequisite that is missing](prereq/missing-prereq.md)

## Validate Prerequisites Ran

If the prerequiste succeeded then the environment variable `PREREQ_RAN` will have been set.

```bash
if [ -n "$PREREQ_RAN" ]; then
    echo "Prerequisites Ran"
else
    echo "FAILURE: Prerequisites did not run"
    exit 1
fi

if [ -n "$VALIDATED_PREREQ_SHOULD_NOT_RUN" ]; then
    echo "FAILURE: Validated prerequisite executed"
    exit 1
else
    echo "Validated prerequisite skipped"
fi

if [ -n "$NESTED_L1_RAN" ]; then
    echo "Nested level 1 prerequisite ran"
else
    echo "FAILURE: Nested level 1 prerequisite did not run"
    exit 1
fi

if [ -n "$NESTED_L2_RAN" ]; then
    echo "Nested level 2 prerequisite ran"
else
    echo "FAILURE: Nested level 2 prerequisite did not run"
    exit 1
fi
```

If succesful you will see:

<!-- expected_similarity="(?s).*Prerequisites Ran.*Validated prerequisite skipped.*Nested level 1 prerequisite ran.*Nested level 2 prerequisite ran.*" -->
```text
Prerequisites Ran
Validated prerequisite skipped
Nested level 1 prerequisite ran
Nested level 2 prerequisite ran
```