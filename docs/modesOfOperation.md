# Modes of Operation

Innovation Engine provides a number of modes of operation. You can view a summary of these with `ie --help`, this document provides more detail about each mode:

  * `execute` - Execute the commands in an executable document without interaction - ideal for unattended execution.
  * `interactive` - Execute a document in interactive mode - ideal for learning.
  * `test` - Execute the commands in a document and test output against the expected. Abort if a test fails. 
  * `to-bash` - Convert the commands in a document into a bash script for standalone execution.
  * `inspect` - Run structural linting and safety checks before executing anything.
  
## Interactive Mode

In Innovation Engine parses the document and presents it one chunk at a time. The the console displays the descriptive text along with the commands to be run and pauses for the user to indicate they are ready to progress. The user can look forward, or backward in the document and can execute the command being displayed (including any outstanding commands up until that point).

This mode is ideal for learning or teaching scenarios as it presents full context and descriptive text. If, however, you would prefer to simply run the commands without interactions use the `execute` mode instead.

## Execute Mode

Execute mode allows for unnatended execution of the document. Unless the script in the document requires user interaction the user can simply leave the script to run in this mode. However, they are also not given the opportunity to review commands before they are executed. If manual review is important use the `interactive` mode instead.

## Test Mode

Test mode runs the commands and then verifies that the output is sufficiently similar to the expected results (recorded in the markdown file) to be considered correct. This mode is similar to `execute` mode but provides more useful output in the event of a test failure.

## To-bash mode

`to-bash` mode does not execute any of the commands, instead is outputs a bash script that can be run independently of Innovation Engine. Generally you will want to send the outputs of this command to a file, e.g. `ie to-bash coolmd > cool.sh`.

## Inspect mode

`inspect` is the guardrail pass you run before executing a scenario in earnest. It parses the entire Markdown file (including prerequisites) and surfaces actionable validation findings without issuing any of the code blocks. Typical problems caught here include:

- Missing descriptive text or language tags on fenced code blocks.
- Prerequisite commands that skip `expected_results` verification (unless the block only contains `export` statements).
- Environment variable issues: lowercase locals are allowed, but uppercase names must either be exported or assigned before use. Unused exports show up as warnings; references to undefined uppercase variables are errors.
- Prefix hygiene: exports must begin with an uppercase prefix (e.g., `PREFIX_VALUE`). `HASH` is the only built-in exception because it is generated automatically for timestamp-safe names.

When `inspect` finds issues it prints both warnings and errors, grouped with counts (for example, `Warning: validation warnings detected (2); see details below.`). If errors exist, the command exits non-zero after reprinting the error summary so CI logs remain readable. Warnings never block execution, but fix them early to keep documents maintainable.

Because `inspect` never runs the commands it is safe to use as a continuous lint pass in your authoring workflow (`ie inspect scenario.md`). Pair it with `ie test` once the structural linting comes back clean.

# Next Steps

<!--
TODO: port relevant content from SimDem to here and update to cover IE
  1. [Hello World Demo](../demo/README.md)
  2. [SimDem Index](../README.md)
  3. [Write SimDem documents](../syntax/README.md)
-->