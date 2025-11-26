# Inline Dependency Prerequisite

This prerequisite is referenced from within another prerequisite section to ensure linked dependencies
are executed even when they are declared inline.

## Setup

```bash
echo "Executing inline dependency"
export INLINE_PREREQ_DEP="dependency-ran"
```

```text
Executing inline dependency
```
