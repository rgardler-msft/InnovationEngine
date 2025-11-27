# Simple Similarity Test, exact match

This command emits a predictable greeting so we can assert similarity handling.

```bash
echo "Hello World"
```

This is what the expected output should be:

<!--expected_similarity=1.0-->

```text
Hello World
```

# Simple Similarity Test, fuzzy match

This command emits a predictable greeting so we can assert similarity handling.

```bash
echo "Hello Jane."
```

This is what the expected output should be:

<!--expected_similarity=0.9-->

```text
Hello Joe
```

# Multi-line code block

Here we span lines with a continuation character to ensure the parser keeps them together.

```bash
echo "Hello \
world"
```

This expected output here is the same, regardless of the multiline code block:

<!--expected_similarity=1.0-->

```text
Hello world
```

# Basic Regex Matching

This block ensures regex expectations can reference exported environment variables.

```bash
export FMT_REGEX_VALUE="RegEx World"
echo "Hello $FMT_REGEX_VALUE"
```

<!-- expected_similarity="^Hello.*"-->

```text
Hello, anyone there?
```

# Regex with ENV variable expansion

This block ensures regex expectations can reference exported environment variables.

```bash
export FMT_REGEX_GREETING="Hello"
export FMT_REGEX_VALUE="RegEx World"
echo "$FMT_REGEX_GREETING $FMT_REGEX_VALUE"
```

<!-- expected_similarity="^$FMT_REGEX_GREETING $FMT_REGEX_VALUE"-->

```text
Hello RegEx World
```

