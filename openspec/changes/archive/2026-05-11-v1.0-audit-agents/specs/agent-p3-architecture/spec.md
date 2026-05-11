# Delta for agent-p3-architecture

## ADDED Requirements

### Requirement: Circuit Breaker Pattern Detection
P3 MUST audit the codebase for circuit breaker patterns at service boundaries. The agent SHALL detect service-to-service calls that lack circuit breaker protection (timeout, failure threshold, half-open recovery). The agent SHALL report calls that would cascade failures under downstream degradation.

(Previously: P3 architecture.md covers coupling, API design, and god objects but has no resilience pattern analysis. This requirement appends a new sub-domain before Calibración de Libertad.)

#### Scenario: Detects missing circuit breaker on external API call
- GIVEN code that calls `http.Post("https://api.payment.com/charge")` without retry or circuit breaker
- WHEN P3 audits service boundaries
- THEN a high-severity finding SHALL be produced
- AND the finding SHALL note that a payment API failure cascades to order creation

#### Scenario: Recognizes existing circuit breaker library usage
- GIVEN Go code using `gobreaker` or `hystrix-go` with configured thresholds
- WHEN P3 audits service boundaries
- THEN no circuit breaker finding SHALL be raised for that call
- AND the existing protection SHALL be noted as a positive indicator

### Requirement: Retry and Timeout Pattern Audit
P3 MUST detect service calls that lack explicit timeouts or retry policies. The agent SHALL flag calls whose default timeouts are unbounded (e.g., Go's `http.Client{}` with zero Timeout). Retries without exponential backoff or jitter SHALL be flagged as insufficient.

#### Scenario: Detects unbounded HTTP client timeout
- GIVEN `http.Client{}` is used without setting `.Timeout`
- WHEN P3 audits network call patterns
- THEN a medium-severity finding SHALL flag the unbounded timeout
- AND the recommendation SHALL suggest a context-based deadline

#### Scenario: Detects retry without backoff
- GIVEN a `for i := 0; i < 3; i++ { call() }` retry pattern without delay
- WHEN P3 audits retry patterns
- THEN a finding SHALL recommend exponential backoff with jitter

### Requirement: Graceful Degradation Assessment
P3 SHALL assess whether the system degrades gracefully when dependencies are unavailable. The agent MUST check for fallback values, cached responses, degraded-mode UI states, and health-check-aware routing.

#### Scenario: Detects missing fallback in feature flag call
- GIVEN code that calls a feature-flag service without a default value
- WHEN P3 audits degradation paths
- THEN a finding SHALL note the hard dependency on feature-flag availability
- AND the recommendation SHALL suggest a sensible default value
