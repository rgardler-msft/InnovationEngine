# An Example Prerequisite: Should need to run

This prerequisite sets the `PREREQ_RAN` variable so that the parent test scenario ([prerequisites.md](prerequisites.md)) can validate it.

## Setup

This step sets the `PREREQ_RAN` environment variable when the prerequisite needs to run.

```bash
export PREREQ_RAN=1
```

## Verification

This verification causes the prerequisite to execute whenever `PREREQ_RAN` has not been set yet.

```bash
if [ -n "$PREREQ_RAN" ]; then
	echo "Prerequisite already executed"
fi

echo "Prerequisite needs to run PREREQ_RAN is not set."
```

If the similarity test fails then the main content of this document will be executed.

<!-- expected_similarity=".*already executed" -->
```text
Prerequisite already executed
```

The validation section will fast-fail. That is, as soon as a failure is detected it is assumed
that the entire document needs to be executed. When the previous test will have failed, the
following will not be executed:

```bash
echo "This test will fail too, but if the first fails this will never run."
```

Giving an output of:

<!-- expected_similarity="Forced failure." -->
```text
This test will fail too, but if the first fails this will never run.
```