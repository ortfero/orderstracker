# orderstracker

tracking market-making order state


## Assumptions

- Unique order identificator. It is assumed that orders are uniquely identified by an OrderClientID.
- Order state consistency. The state machine controlling order transitions (OrderUnplaced, OrderPlacing, OrderPlaced, OrderModifying, OrderCanceling and OrderFilled) assumes that appropriate functions are called by exchange gateway.
- In addition to the order status, we store the last execution report. This allows us to recognize different corner cases. For example, the order was placed, but an attempt to modify its price later failed. In this case, the order status will stay 'OrderPlaced' but the execution report will be 'ReportRejected'.


## Trade-offs

- Simple data structures. The implementation uses nested maps (for exchanges and symbols) to organize market data. The alternative would be to use composable key 'exchange+symbol' with flat map but it implies allocation for every key search.
- Thread safety via a global mutex. There is an implicit belief that the overhead of a global lock is acceptable relative to its simplicity. The alternative would be to use concurrent map or event-driven architecture with channels.
- Aggregation via VWAP. For partially filled orders, the code aggregates executions using a basic Volume Weighted Average Price (VWAP) calculation. We are loosing more granular trade details but reducing memory overhead. It assumes that such aggregation is acceptable for the application's requirements. The alternative would be to store each trade


## Source code

- `order.go` -- data types for exchange, symbol, information about order and order status
- `executionreport.go` -- information about the status of the last order action
- `tracker.go` -- data types and functions to track orders status

## Run tests

```shell
go test -v
```

## Run benchmarks

```shell
go test -bench=.
```

## Benchmarks

12.04.2025
```
cpu: AMD Ryzen 7 8845HS w/ Radeon 780M Graphics
BenchmarkTracker_OrderGenerateAndPlace-16        5600036               210.2 ns/op
```
