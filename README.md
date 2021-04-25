# Distributed Systems Chubby Lock

## Contributors
-Shermine Chua Xin Min 
-Tan Yu Xiang 
-Wang Zefan 
-Zhang Jingyu 
-Akmal Hakim Teo 

## Introduction


In a distributed system, data is duplicated mainly for reliability and performance. Replication is required, especially if the distributed system needs to grow in quantity and geographically. Consistency within a distributed system has been an ongoing issue which creates inconvenience for service providers and their customers. There is a need for nodes within a distributed system to agree on information and to share data reliably. In reality, we cannot forgo the consistency constraints completely, thus it is necessary to understand how much consistency is required in the system and how to make full use of it.

## Problem Formulation
How might we provide a lock service for a distributed system that ensures safe resource sharing and concurrency? The method that we will be referring to, is the Chubby lock service. The Chubby lock service aims to provide coarse-grained locking that focuses mainly on availability and reliability over high-performance. Similarly, we aim to implement the following key features:

1) Fault Tolerance
2) Availability
3) Safety
4) Sequential Consistency

## System Overview
Our system consists of a Chubby cell and client processes. A Chubby cell consists of five servers, each with their own database, provided by BoltDB. Within the Chubby cell, one of the servers (with the highest node ID) will be elected as the master. The rest of the servers in the Chubby cell serve as replicas in order to reduce the likelihood of correlated failure. The replicas maintain a copy of the master’s database in their own database and will not initiate any reads/ writes from the clients.

All client reads/ writes will have to go through the master server, hence the client will have to locate the master server in the Chubby cell to proceed with any requests. This is piggybacked on the KeepAlive messages that are periodically sent between nodes. Once a client locates the master, it will direct all requests to the master until the master ceases to respond or it has been notified of a master change by the other servers. Write requests are propagated to all replicas and will only proceed on the master database when all the four replicas in the Chubby cell acknowledge the update on their database. Read requests are also handled by the master and master will return the requested data from its own database. In case of a master failure, the other replicas will run the Bully Algorithm to elect a new master and in the meantime, all requests from the clients will be blocked. A write request from the client will fail as long as a single server in the Chubby cell is down as a successful write requires acknowledgement from all 4 replicas and the master.


## Instructions

### Requirements:
1. Docker
2. Golang

### How to run normally
1. `cd` into the directory
2.  run `docker-compose build`
**NOTE**: Whenever you change any of the code, remember to run the build command first. To decide on the number of nodes to run, change the docker-compose.yml file by commenting/uncommenting the nodes that you want.
4.  run `docker-compose up`


### How to simulate multiple clients per client dockerfile requesting for same file
Currently, each client dockerfile represents 1 client. To simulate multiple clients requesting for the same file, follow these instructions:
1. In each Dockerfile, comment out line 13
2. In each Dockerfile, uncomment line 12 
**NOTE:**: For `CMD ["./multiple_client.sh", "30", "3", "master"]` 
0th argument `"./multiple_client.sh""` contains the bash script simulating the scalability test case.
1st argument `"30"` refers to the number of clients you wish to simulate PER dockerfile. You can change the 1st argument to match the number of clients you wish to simulate.
2nd argument `0` follows the dockerfile client id
3. run `docker-compose build`
4. run `docker-compose up`

### How to simulate multiple clients per client dockerfile requesting for different file
Currently, each client dockerfile represents 1 client. To simulate multiple clients requesting for the same file, follow these instructions:
1. In each Dockerfile, comment out line 13
2. In each Dockerfile, uncomment line 10
**NOTE:**: For `CMD ["./diff_file.sh", "30", "0", "master"]` 
0th argument `"./diff_file.sh"` contains the bash script simulating the scalability test case.
1st argument `"30"` refers to the number of clients you wish to simulate PER dockerfile. You can change the 1st argument to match the number of clients you wish to simulate.
2nd argument `0` follows the dockerfile client id
3. run `docker-compose build`
4. run `docker-compose up`



