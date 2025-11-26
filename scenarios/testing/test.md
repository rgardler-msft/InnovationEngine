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

## Validation

This final check confirms that the aggregated scenarios completed successfully.

```bash
# Final sanity check after all test scenarios have run.
echo "Master test scenario completed."
```
