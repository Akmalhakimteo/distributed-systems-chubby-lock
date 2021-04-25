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


### Experiments, Evaluations, Tests

Bully’s Algorithm




Fig 4: Bully’s Algo Test case
As shown in the diagrams above, if server 4 (master server) fails at any time, the remaining replicas start election to elect a new master server who has the highest node id, which in this case, server 3 gets elected. It will then propagate its database contents and exchange keepalives with the clients.

To reproduce test results:

After executing the program, kill the master server manually in docker desktop by stopping it. Expected outcome should be as similar to above mentioned results.

Master Server Failure
To ensure that when the master server fails, the remaining replicas start the election to elect a new master server who has the highest node id. Any ongoing write requests will fail, locks will be released and clients will have to request to write again to the newly elected master. Furthermore, the newly elected master server will propagate its database to all replicas servers.





Fig 5: Master server failure test case
As shown in the figures above, server 4 (master server) fails in the middle of a write request. Other replica servers that found out about this start election to be the new master server simultaneously and server 3 eventually wins to be the new master server. It then propagates its database contents to all remaining replica servers to ensure consistency.

To reproduce test results:

First ensure that read and write calls in rpc_client.go are not commented.

After executing the program, kill the master server manually in docker desktop by stopping it during a write request. Expected outcome should be as similar to above mentioned results.

Replica Failure
To ensure that when a server replica fails, because there are less than five alive servers in the chubby cell, write requests by clients are failed and a rollback event occurs to ensure consistency in server databases.




Fig 6: Replica server failure test case
As shown in the figures above, server 3 (replica server) fails when server 4 (master server) is serving the write request. As such, server 4 will not receive 4 ACKs from all the replica servers in the chubby cell. It will then fail the write request and propagate the current copy (not updated) of its database across all replica servers to ensure consistency.

To reproduce test results:

First ensure that read and write calls in rpc_client.go are not commented.

After executing the program, kill any replica server manually in docker desktop by stopping it during write propagating as indicated by the print statement: “Server: 4 is forwarding write to other servers”. Expected outcome should be as similar to above mentioned results.

Client Failure
To ensure that when a client fails during a write request, the lock to the file is automatically released by the Master Server.




Fig 7: Client failure test case
The figures above show that when client 1 fails during a write request, the write request is still served and that the lock will be released by the master server automatically to allow other clients to acquire.

To reproduce test results:

First ensure that read and write calls in rpc_client.go are not commented.

After the client has been granted the lock after approximately 10 seconds, manually kill the client in docker desktop to observe the above mentioned results.


Master Server Recovery
To ensure that when a server with higher node id wakes up, it wins the election and receives the database contents of the current master server.

Fig 8: Master server recovery test case
Figure 8 shows that when Server 4 wakes up, it first gets the current coordinator id, starts election and then wins the election to be the new master server. After which, it will receive the database contents of the old master server which is server 3.

To reproduce test results:

After executing the program, kill the master server manually in docker desktop by stopping it. 
After a few seconds, manually restart the server by starting it again. Expected outcome should be as similar to above mentioned results.

Keepalives
To conduct constant health checks on the status of clients and servers.

Fig 9: Keepalives test case

Figure 9 shows that Keepalives messages are constantly exchanged between the clients and the master server, which in this case, is Server 4. Chubby servers will also exchange ping messages among themselves to check on their individual status. In an event of master server failure, chubby servers will begin election until the server with highest id wins to be the new master server, and the clients will exchange keepalives with the new master server.

To reproduce test results:
Execute the program normally to observe keepalive exchanges as shown above in server/client containers.

Safety
To ensure that only one client can be granted the lock to a file.

Fig 10.1: Safety test case
Figure 10.1 shows that Client 2 has been granted the lock and Server 4 is currently propagating the write request to other replicas.


Fig 10.2: Safetytest case
Figure 10.2 shows that while the write request from Client 2 is still in progress, Client 4 tries to acquire the lock and fails because Client 2 is still holding on to the lock.
This ensures that, at any one time, only one client can be granted the lock to a specific file.

To reproduce test results:

First ensure that read and write calls in rpc_client.go are not commented.
Execute the program normally to observe multiple clients trying to write to the same file. Expected outcome should be as similar to above mentioned results.

Availability
To test if other clients can still read the file when a write request is being served.

Fig 11: Availability test case

Figure 11 above shows that even read requests from multiple clients can be served concurrently while a write request is being served. The clients are reading “key value that does not exist“ because nothing has been written by the first client which has obtained the lock.




