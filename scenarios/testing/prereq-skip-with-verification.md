# A Prerequisite Example: Should not need to run

This prerequisite demonstrates how a verification section can prevent the remainder of the prerequisite from running when the environment is already in a good state.

## Setup

If this is run we will set an environment variable. This allows us to test that it wasn't run, which is as expected.

```bash
echo "Validated prerequisite body executing"
export VALIDATED_PREREQ_SHOULD_NOT_RUN=1
```

## Verification

This will always pass and thus this prerequisite should never run.

```bash
echo "Validated prerequisite already satisfied"
```

<!-- expected_similarity = 1.0 -->
```text
Validated prerequisite already satisfied
```
