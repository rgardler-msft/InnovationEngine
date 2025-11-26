# Variable declaration and usage

## Simple declaration

Set and immediately echo a single environment variable.

```bash
export MY_VAR="Hello, World!"
echo $MY_VAR
```

Expected output:

<!-- expected_similarity=1.0 -->

```text
Hello, World!
```

## Double variable declaration

Declare two variables on one line while only echoing the first.

```bash
export NEXT_VAR="Hello" && export OTHER_VAR="Hello, World!"
echo $NEXT_VAR
```

Expected output:

<!-- expected_similarity=1.0 -->

```text
Hello
```

## Double declaration with semicolon

Use semicolons to chain exports before reading the second variable.

```bash
export THIS_VAR="Hello"; export THAT_VAR="Hello, World!"
echo $THAT_VAR
```

Expected output:

<!-- expected_similarity=1.0 -->

```text
Hello, World!
```

## Declaration with subshell value

Capture a subshell result into an environment variable.

```bash
export SUBSHELL_VARIABLE=$(echo "Hello, World!")
echo $SUBSHELL_VARIABLE
```

Expected output:

<!-- expected_similarity=1.0 -->

```text
Hello, World!
```

## Declaration with other variable in value

Build one variable from another and display the composed result.

```bash
export VAR1="Hello"
export VAR2="$VAR1, World!"
echo $VAR2
```

Expected output:

<!-- expected_similarity=1.0 -->

```text
Hello, World!
```
