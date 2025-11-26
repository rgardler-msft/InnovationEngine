# Test Prerequisites

This document has both present and missing prerequisites. The one that is present should be executed, allowing this document to fully execute. The missing one should only produce a warning and not abort execution.

## Prerequisites

The following prerequisite documents will be executed before validation. One exists, one does not to exercise warning + continue behavior, one includes verification logic that should cause it to skip execution, and one exercises multi-level nested prerequisites.

- [A prerequisite that needs to run](prereq/prereq-set-variable.md)
- [A prereuisite that does not need to run](prereq/prereq-skip-with-verification.md)
- [A nested prerequisite chain](prereq/prereq-nested-level1.md)
- [A prerequisite that is missing](prereq/missing-prereq.md)
- [A prerequisite with inline content](prereq/prereq-inline-section.md)

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

if [ "$INLINE_PREREQ_LOG" = "stage1:stage2:stage3" ]; then
    echo "Inline prerequisite section commands executed"
else
    echo "FAILURE: Inline prerequisite section did not execute in order"
    exit 1
fi

if [ "$INLINE_PREREQ_DEP" = "dependency-ran" ]; then
    echo "Inline dependency executed"
else
    echo "FAILURE: Inline dependency did not execute"
    exit 1
fi
```

If succesful you will see:

<!-- expected_similarity=0.9 -->
```text
Prerequisites Ran
Validated prerequisite skipped
Nested level 1 prerequisite ran
Nested level 2 prerequisite ran
Inline prerequisite section commands executed
Inline dependency executed
```