# Master Test Scenario

This master test scenario runs all existing test documents in this folder as prerequisites so they can be exercised together.

## Prerequisites

- [Code blocks execution](CodeBlocks.md)
- [Comments handling](CommentTest.md)
- [Fuzzy matching](fuzzyMatchTest.md)
- [Prerequisites aggregation test](prerequisites.md)
- [Reporting and expectations](reporting.md)
- [Streaming output](test-streaming.md)
- [Variable hierarchy scenario](variableHierarchy.md)
- [Variables basics](variables.md)

## Steps

Each of the following steps are dummy steps to allow full testing of reporting and linting tools.

### Step 1: Hello

This step says hello.

```bash
echo "Hello!"
```

### Step 2: Sleep

Wait for a couple of seconds.

```bash
sleep 2
```

### Step 3: Goodbye

This step says Goodbye.

```bash
echo "Goodbye".
```

## Validation

This final check confirms that the aggregated scenarios completed successfully.

```bash
# Final sanity check after all test scenarios have run.
echo "Master test scenario completed."
```
