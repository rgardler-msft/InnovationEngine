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

<!-- expected_similarity="Prerequisite already executed" -->
```text
Prerequisite already executed
```

