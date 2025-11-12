# Test Streaming Output

This scenario tests real-time output streaming.

## Step 1: Quick output

This should appear immediately:

```bash
echo "Line 1"
echo "Line 2"
echo "Line 3"
```

## Step 2: Delayed output

This will show output as it's generated (one line per second):

```bash
for i in {1..5}; do
  echo "Progress: $i/5"
  sleep 1
done
echo "Complete!"
```

## Step 3: Continuous output

This generates output over several seconds:

```bash
echo "Starting long-running task..."
for i in {1..10}; do
  echo "  Processing item $i..."
  sleep 0.5
done
echo "Task finished!"
```

## Step 4: Multiple commands with delays

```bash
echo "=== Phase 1 ==="
sleep 1
echo "Phase 1 complete"
echo ""
echo "=== Phase 2 ==="
sleep 1
echo "Phase 2 complete"
echo ""
echo "=== Phase 3 ==="
sleep 1
echo "Phase 3 complete"
```
