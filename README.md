# Prototype PSI/oo-PIR hybrid strategy

We prototype the PSI scheme detailed in [Scaling Mobile Private Contact Discovery to Billions of Users (2023)](https://eprint.iacr.org/2023/758.pdf). This code was written to gauge server costs of using such a scheme. It leaves out details from the paper and we cannot guarantee full security. Furthermore, the code only support clients running on the same machine as the server. 

We also include an implementation of Checklist PIR on a partitioned database (DB-PIR).

## Usage
This repo includes code for two different experiments. Communication experiments output the communication and time taken to run the protocol on a single client. Throughput experiments measure the time taken to serve 10 clients concurrently.

### Communication Experiments
We include examples for two different methods:

PIR/oo-PIR hybird strategy at [checklist/example/psi_pir_hybrid.go](checklist/example/psi_pir_hybrid.go):
```
go run checklist/example/psi_pir_hybrid.go
```

DB-PIR at [checklist/example/db_pir.go](checklist/example/db_pir.go):
```
go run checklist/example/db_pir.go
```

For each number of partitions specified in the code, the following will be printed:
- Total bits for server to send hints
  - This is printed for each partition, so total offline communication = (num bits to send hints * num partitions)
- Offline time (Hint generation:)
- Total bits for each online client query and server to answer each query
- Number of client requests and server responses (at the bottom)
  - Total online communicaiton = (num bits for online query * num client requests + num bits for server to answer query * num server responses)
 - Online time (Online phase:)


### Throughput Experiments
To run multiple clients concurrently:

PIR/oo-PIR hybird strategy at [checklist/example/threaded_psi_pir.go](checklist/example/threaded_psi_pir.go):
```
go run checklist/example/threaded_psi_pir.go
```

DB-PIR at [checklist/example/threaded_db_pir.go](checklist/example/threaded_db_pir.go):
```
go run checklist/example/threaded_db_pir.go
```

These commands will print the time taken to answer all client queries.


## References
We use the following open source repos in our implementation:
- [Cuckoo Filter](https://github.com/irfansharif/cfilter)
- [Checklist PIR](https://github.com/dimakogan/checklist), from the paper [Mobile Private Contact Discovery at Scale (2019)](https://www.usenix.org/system/files/sec19-kales.pdf)

## Support 
Email kameronshahabi@gmail.com with any questions.
