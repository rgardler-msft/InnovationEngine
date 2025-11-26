# Testing multi Line code block

This command deliberately emits a predictable greeting so we can assert similarity handling.

```azurecli-interactive
echo "Hello World"
```

This is what the expected output should be

<!--expected_similarity=0.8-->

```text
Hello world
```

# Testing multi Line code block

Here we span lines with a continuation character to ensure the parser keeps them together.

```azurecli-interactive
echo "Hello \
world"
```

# Output Should Fail

This expected output intentionally mismatches the command above.

<!--expected_similarity=0.9-->

```text
Hello world
```

# Code block

We repeat the multiline command to exercise additional comparisons.

```azurecli-interactive
echo "Hello \
world"
```

# Output Should Pass

This expected output mirrors the command exactly so similarity should succeed.

<!--expected_similarity=1.0-->

```text
Hello world
```

# Code block

One more repeated command keeps the test matrix simple.

```azurecli-interactive
echo "Hello \
world"
```

# Bad similarity - should fail

This final expectation purposefully uses a failing similarity threshold.

<!--expected_similarity=0.9-->

```text
Hello world
```
