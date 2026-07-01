run file seed_products.sql: docker exec -i product-trace-db psql -U postgres -d product_trace_db -f - < infra\postgres\seed\seed_products.sql > seed.log 2>&1

run file seed_traceability.sql: docker exec -i product-trace-db psql -U postgres -d product_trace_db -f - < infra\postgres\seed\seed_traceability.sql > seed_trace.log 2>&1