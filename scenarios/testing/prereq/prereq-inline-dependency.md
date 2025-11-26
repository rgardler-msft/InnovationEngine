# Inline Dependency Prerequisite

This prerequisite is referenced from within another prerequisite section to ensure linked dependencies
are executed even when they are declared inline.

## Setup

Run the inline dependency commands so downstream prerequisites can detect completion.

```bash
echo "Executing inline dependency"
export INLINE_PREREQ_DEP="dependency-ran"
```

Expected output for the dependency command:

```text
Executing inline dependency
```
