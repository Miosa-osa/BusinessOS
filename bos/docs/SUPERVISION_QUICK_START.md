# Supervision Tree — Quick Start Guide

## In 60 Seconds

```rust
use bos_core::{SupervisorHandle, SupervisorConfig, WorkerConfig};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // 1. Create supervisor
    let supervisor = SupervisorHandle::new(SupervisorConfig::root());

    // 2. Add workers (Discovery: 5min, Conformance: 30sec, IO: 10sec)
    let discovery = supervisor.add_worker(WorkerConfig::discovery()).await?;

    // 3. Define work (restarts automatically on crash)
    let work = || Box::pin(async {
        println!("Working...");
        tokio::time::sleep(tokio::time::Duration::from_secs(1)).await;
        Ok::<(), _>(())
    });

    // 4. Supervise (automatic crash detection + restart)
    supervisor.supervise_worker(discovery, work).await?;

    // 5. Wait for completion
    supervisor.wait_all().await?;

    Ok(())
}
```

**Result:** Worker crashes are detected, restarted with exponential backoff (100ms → 200ms → 400ms → 800ms), with max 5 restarts in 60 seconds.

## Key Concepts

### Process State Machine

```
   START
     ↓
  [Starting] → [Running]
     ↑           ↓
     └── [Restarting] ← [Crashed]
                        (on error)
     [Shutting] → [Stopped]
```

### Restart Policy (Default)

```
Strategy:         OneForOne (restart only failed)
Max restarts:     5
Time window:      60 seconds
Initial backoff:  100ms
Max backoff:      800ms
Exponential:      yes (doubles each restart)
```

If worker crashes > 5 times in 60 seconds → give up (propagate error).

### Three Restart Strategies

| Strategy | Behavior | Use When |
|----------|----------|----------|
| **OneForOne** | Restart only failed worker | Workers independent |
| **OneForAll** | Restart failed + all siblings | Workers tightly coupled |
| **RestForOne** | Restart failed + started-after | Dependency chain |

## Common Patterns

### Pattern 1: Independent Discovery Worker

```rust
let supervisor = SupervisorHandle::new(SupervisorConfig::root());
let discovery = supervisor.add_worker(WorkerConfig::discovery()).await?;

supervisor.supervise_worker(discovery, || {
    Box::pin(async {
        let model = discover_process_model().await?;
        println!("Model: {:?}", model);
        Ok(())
    })
}).await?;
```

**Restarts automatically if `discover_process_model()` fails.**

### Pattern 2: State Recovery

```rust
let worker = supervisor.add_worker(WorkerConfig::conformance()).await?;

// Before critical work
let checkpoint = vec![1, 2, 3, 4, 5];
worker.save_state(checkpoint).await?;

// After crash, resume
if let Some(saved) = worker.restore_state().await? {
    let position = saved[0] as usize;
    println!("Resuming from position {}", position);
}
```

**Prevents duplicate work on restart.**

### Pattern 3: Multi-Worker Supervision

```rust
let supervisor = SupervisorHandle::new(SupervisorConfig::root());

// Add 3 workers with different timeouts
let discovery = supervisor.add_worker(WorkerConfig::discovery()).await?;    // 5 min
let conformance = supervisor.add_worker(WorkerConfig::conformance()).await?;  // 30 sec
let io = supervisor.add_worker(WorkerConfig::io()).await?;                    // 10 sec

// Supervise independently
supervisor.supervise_worker(discovery, discovery_fn).await?;
supervisor.supervise_worker(conformance, conformance_fn).await?;
supervisor.supervise_worker(io, io_fn).await?;

// All restart independently on crash
supervisor.wait_all().await?;
```

### Pattern 4: Custom Restart Policy

```rust
let mut policy = RestartPolicy::default_one_for_one();
policy.max_restarts = 10;           // Allow more attempts
policy.time_window_secs = 120;      // In 2-minute window
policy.initial_backoff_ms = 50;     // Start faster
policy.exponential_backoff = true;  // But exponentially

let mut config = SupervisorConfig::root();
config.restart_policy = policy;
let supervisor = SupervisorHandle::new(config);
```

### Pattern 5: Timeout Handling

```rust
// Create worker with custom timeout
let mut config = WorkerConfig::io();
config.timeout_ms = 5000;  // 5 second timeout
let worker = supervisor.add_worker(config).await?;

// Work function that might timeout
supervisor.supervise_worker(worker, || {
    Box::pin(async {
        // If this takes > 5 seconds, supervisor cancels and restarts
        tokio::time::sleep(tokio::time::Duration::from_secs(10)).await;
        Ok(())
    })
}).await?;
```

## Configuration Reference

### WorkerConfig Presets

```rust
// Discovery: 5 minute timeout, 100k markings max
let config = WorkerConfig::discovery();

// Conformance: 30 second timeout
let config = WorkerConfig::conformance();

// I/O: 10 second timeout
let config = WorkerConfig::io();

// Custom
let mut config = WorkerConfig::new("MyWorker", 15000);  // 15 seconds
config.memory_limit_bytes = 4 * 1024 * 1024 * 1024;     // 4GB
config.max_markings = 50_000;
```

### RestartPolicy Presets

```rust
// Default: OneForOne, 5 in 60s, 100ms-800ms exponential backoff
let policy = RestartPolicy::default_one_for_one();

// Strict: OneForAll, 3 in 60s
let policy = RestartPolicy::strict_one_for_all();

// Custom
let mut policy = RestartPolicy {
    strategy: RestartStrategy::RestForOne,
    max_restarts: 10,
    time_window_secs: 120,
    initial_backoff_ms: 50,
    max_backoff_ms: 5000,
    exponential_backoff: true,
};
```

## Error Handling

```rust
match result {
    Ok(()) => println!("Success"),
    Err(SupervisionError::RestartLimitExceeded { max_restarts, window_secs }) => {
        eprintln!("Too many restarts: {} in {}s", max_restarts, window_secs);
    }
    Err(SupervisionError::WorkerTimeout { name, timeout_secs }) => {
        eprintln!("Worker {} timeout after {}s", name, timeout_secs);
    }
    Err(SupervisionError::MemoryLimitExceeded { current_bytes, limit_bytes }) => {
        eprintln!("Memory: {} > {}", current_bytes, limit_bytes);
    }
    Err(e) => eprintln!("Error: {}", e),
}
```

## Testing Your Workers

```rust
#[tokio::test]
async fn test_discovery_recovery() {
    let supervisor = SupervisorHandle::new(SupervisorConfig::root());
    let worker = supervisor.add_worker(WorkerConfig::discovery()).await.unwrap();

    // Simulate crash
    worker.set_state(WorkerState::Crashed).await;
    assert_eq!(worker.state().await, WorkerState::Crashed);

    // Simulate recovery
    worker.set_state(WorkerState::Restarting).await;
    worker.set_state(WorkerState::Running).await;
    assert_eq!(worker.state().await, WorkerState::Running);
}
```

## Backoff Timing

Default exponential backoff prevents crash loops:

```
Crash 1: wait 100ms → restart
Crash 2: wait 200ms → restart
Crash 3: wait 400ms → restart
Crash 4: wait 800ms → restart
Crash 5: wait 800ms → restart
Crash 6: GIVE UP (error: RestartLimitExceeded)
```

## State Diagram

```
supervisor.supervise_worker()
    ↓
[Starting] ← user creates worker
    ↓
[Running] ← work_fn runs
    ↓
┌─────────────────────┬──────────────────┐
│ (Success)           │ (Error)          │
│ work_fn returns Ok  │ work_fn returns Err
│       ↓             │       ↓
│  [Stopped]          │   [Crashed]
│                     │       ↓
│                     │ record_restart()
│                     │       ↓
│                     │  [Restarting]
│                     │    + sleep(backoff)
│                     │       ↓
│                     │  [Running] ← loop
│                     │       (back to line 3)
│                     │
└─────────────────────┘

If restarts exceed limit:
[Crashed] → (RestartLimitExceeded error)
```

## Checklist: Adding a Supervised Worker

- [ ] Create `WorkerConfig` (use preset or custom)
- [ ] Add to supervisor with `add_worker()`
- [ ] Define work function (async closure)
- [ ] Call `supervise_worker()` with work function
- [ ] (Optional) Implement state save/restore
- [ ] Test with intentional crashes
- [ ] Monitor restart metrics in production

## Next Steps

1. **Read Full Docs:** `/Users/sac/chatmangpt/BusinessOS/bos/docs/SUPERVISION_TREE_ARCHITECTURE.md`
2. **Run Tests:** `cargo test --test supervision_integration_test`
3. **Study Examples:** Look at `core/tests/supervision_integration_test.rs`
4. **Integrate:** Add supervision to your worker processes
