# bpfstream

Fast bpftrace event stream processor, write records to DuckDB

## Benchmark

```
linux-amd64, cpu: AMD Ryzen 7 9700X 8-Core Processor

BenchmarkImportFromBpf
elapsed_ns=1104481628/4 rows=321039
-> ns_per_row=860, 1240694 ops
```
