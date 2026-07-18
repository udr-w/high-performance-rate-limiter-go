# Security Policy

## Supported Versions

Security fixes are provided for the latest commit on the `main` branch until the project starts publishing tagged releases.

When tagged releases are introduced, this policy should be updated with an explicit support matrix.

## Reporting a Vulnerability

Please do not open a public GitHub issue for a suspected vulnerability.

Report security concerns privately by emailing the maintainer:

- `w.udara@yahoo.com`

Include as much detail as possible:

- Affected package, version, commit, or configuration.
- Clear reproduction steps or proof of concept.
- Expected and observed behavior.
- Impact assessment, including whether the issue can cause bypass, denial of service, data exposure, or unsafe Redis behavior.
- Any logs, stack traces, benchmark output, or Redis state that helps reproduce the issue.

## Response Expectations

The maintainer will aim to:

- Acknowledge the report within 7 days.
- Triage severity and reproducibility within 14 days.
- Provide status updates when a fix needs more time.
- Credit reporters when requested and appropriate.

Confirmed vulnerabilities will be fixed privately where practical, then disclosed after a patched commit or release is available.

## Security Scope

Security-sensitive areas include:

- Rate-limit bypasses under concurrent access.
- Redis atomicity, TTL, failover, timeout, and unavailable-service behavior.
- Unbounded Redis key growth or unsafe handling of untrusted identifiers.
- Data races in limiter implementations.
- Configuration behavior that unexpectedly fails open or fails closed.
- Denial-of-service behavior caused by attacker-controlled inputs.

## Secure Usage Guidance

- Prefer constructors that return errors, such as `NewTokenBucketWithOptions`, `NewLeakyBucketWithOptions`, and `NewRedisLimiterWithValidation`.
- Use `RedisLimiter.AllowContext` with request-scoped deadlines so Redis latency cannot consume unbounded request time.
- Decide and document your application policy for Redis errors: fail closed, fail open, or use a local fallback limiter.
- Use `RedisKey` or an equivalent namespaced and hashed key scheme for untrusted identifiers.
- Configure Redis authentication, ACLs, network restrictions, timeouts, memory limits, and eviction policies for limiter data.
- Treat the Redis limiter as fixed-window. If your threat model cannot tolerate boundary bursts, use or implement a sliding-window or token-bucket distributed limiter.

## Out of Scope

The following are usually out of scope unless they demonstrate a concrete vulnerability in this library:

- Issues caused only by an insecure Redis deployment.
- Vulnerabilities in downstream applications that use the library incorrectly.
- Denial-of-service reports without a practical reproduction or impact explanation.
- Reports against unsupported forks or modified versions.
