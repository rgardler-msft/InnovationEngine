<!--This is a test comment in markdown -->

<!--This is a multi line comment in markdown

The text in here should not show up

 in markdown -->

# Testing multi Line code block

This command spans two shell lines so we can verify our parsing handles escaped newlines.

```bash
echo "Hello \
world"
```

# This is what the output should be

The output block below is referenced by similarity validation comments.

<!--expected_similarity=0.8-->

```text
hello world
```
