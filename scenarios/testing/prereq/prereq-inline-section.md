# Inline Section Prerequisite

This document exercises inline prerequisite content. Commands inside the `## Prerequisite` section,
including nested headings, should execute before the primary steps of the parent document. It also
references another prerequisite from within that section to validate execution order for dependencies.

## Prerequisite

### Seed Inline Log

```bash
export INLINE_PREREQ_LOG="stage1"
```

### Nested Inline Stage

The following nested heading ensures subsections are treated as part of the prerequisite.

#### Append Stage Two

```bash
export INLINE_PREREQ_LOG="${INLINE_PREREQ_LOG}:stage2"
```

### Linked Dependency

This prerequisite also depends on an additional inline dependency. The dependency must execute when it
is discovered.

- [Inline dependency](prereq-inline-dependency.md)

### Final Inline Stage

```bash
export INLINE_PREREQ_LOG="${INLINE_PREREQ_LOG}:stage3"
```

## Steps

```bash
echo "Inline prerequisite body complete"
```
