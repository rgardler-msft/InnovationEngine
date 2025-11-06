# Test Prerequisites

This document has both present and missing prerequisites. The one that is present should be executed, allowing this document to fully execute. The missing one should only produce a warning and not abort execution.

## Prerequisites

The following prerequisite documents will be executed before validation. One exists, one does not to exercise warning + continue behavior.

- [Present prerequisite](prereq-set-variable.md)
- [Missing prerequisite](missing-prereq.md)

## Validate Prerequisites Ran

If the prerequiste succeeded then the environment variable `PREREQ_RAN` will have been set.

```bash
if [ -n "$PREREQ_RAN" ]; then
    echo "Prerequisites Ran"
else
    echo "FAILURE: Prerequisites did not run"
fi
```

If succesful you will see:

<!-- expected_similarity=".*Ran.*" -->
```text
Prerequisites Ran
```