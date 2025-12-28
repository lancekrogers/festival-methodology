# Performance Check

**Gate:** performance_check | **Status:** Pending

## Objective

Verify implementation meets performance requirements and identify optimization opportunities.

## Performance Checklist

### Response Time

- [ ] API endpoints respond within acceptable limits
- [ ] Database queries are optimized
- [ ] Caching implemented where beneficial
- [ ] Lazy loading used for large data sets

### Resource Usage

- [ ] Memory usage is within bounds
- [ ] No memory leaks identified
- [ ] CPU usage is acceptable
- [ ] File handles and connections properly closed

### Scalability

- [ ] Code handles concurrent requests properly
- [ ] Rate limiting implemented where needed
- [ ] Connection pooling used for database
- [ ] Horizontal scaling considered in design

### Benchmarks

- [ ] Performance tests written
- [ ] Baseline metrics established
- [ ] Regression tests pass

## Measurements

<!-- Document key performance metrics -->

| Metric | Target | Actual | Status |
|--------|--------|--------|--------|
| p95 Response Time | <100ms | | |
| Memory Usage | <256MB | | |
| Throughput | >1000 req/s | | |

## Optimization Notes

<!-- Document any optimizations made -->

## Sign-off

- [ ] Performance review completed
- [ ] Meets requirements
- [ ] Ready for production load
