# surfstore

## overview

<img width="1008" alt="iShot_2023-02-20_17 51 45" src="https://user-images.githubusercontent.com/114261503/220230310-b0c30799-3f71-4d93-8b90-3f16432ed911.png">


This is a cloud-based file storage service called SurfStore. SurfStore is a networked file storage application that is based on Dropbox, and lets you sync files to and from the “cloud”. Cloud services are implemented, and a client is designed which interacts with the service via gRPC.

Multiple clients can concurrently connect to the SurfStore service to access a common, shared set of files. Clients accessing SurfStore “see” a consistent set of updates to files, but SurfStore does not offer any guarantees about operations across files, meaning that it does not support multi-file transactions (such as atomic move).

## Fundamentals

### Blocks, hashes, and hashlists

A file in SurfStore is broken into an ordered sequence of one or more blocks. Each block is of uniform size (defined by the command line argument), except for the last block in the file, which may be smaller (but must be at least 1 byte large). As an example, assume the block size is 4096 bytes, and consider the following file:

<img width="1149" alt="iShot_2023-02-20_17 55 35" src="https://user-images.githubusercontent.com/114261503/220230353-bbf67291-f017-43e7-96ca-67ec0d04015a.png">


### The Base Directory

A command-line argument specifies a “base directory” for the client. This is the directory that is going to be synchronized with your cloud-based service. Your client will upload files from this base directory to the cloud, and download files (and changes to files) from the cloud into this base directory. Your client should not modify any files outside of this base directory. Note in particular–your client should not download files into the “current” directory, only the base directory specified by that command line argument. 

### Version

Each file/filename is associated with a version, which is a monotonically increasing positive integer. The version is incremented any time the file is created, modified, or deleted. The purpose of the version is so that clients can detect when they have an out-of-date view of the file hierarchy.
Note: if two clients try to modify the same file in the cloud concurently, the one that uploads the change later will recieved a "upload failure" flag.

<img width="1140" alt="iShot_2023-02-20_18 15 39" src="https://user-images.githubusercontent.com/114261503/220230681-02cd3e17-b6da-4987-bc6e-19ed9be2385a.png">

## Usage Details

### Client

clients will sync the contents of a “base directory” by:

```
go run cmd/SurfstoreClientExec/main.go -d <meta_addr:port> <base_dir> <block_size>

Usage:
-d:
Output log statements
<meta_addr:port>: 
(required) IP address and port of the MetaStore the client is syncing to
<base_dir>:
(required) Base directory of the client 
<block_size>:
(required) Size of the blocks used to fragment files



```
note: index.db will be maintained automatically, don't modify this file manually.
### Server 

Surfstore is composed of two services: MetaStore and BlockStore. the location of the MetaStore and BlockStore doen't matter at all. In other words, the MetaStore and BlockStore could be serviced by a single server process, separate server processes on the same host, or separate server processes on different hosts. Regardless of where these services reside, the functionality is always the same.

```
go run cmd/SurfstoreServerExec/main.go -s <service_type> -p <port> -l -d (blockstoreAddrs*)

Usage:
-s <service_type>: 
(required) This defines the service provided by this server. It can be “meta”, “block”, or “both” (you don’t need to include the quotation marks).
-p <port>:
(default=8080) Port to accept connections
-l:
Only listen on localhost if included
-d:
Output log statements
(blockStoreAddrs*):
BlockStore addresses (ip:port) the MetaStore should be initialized with. Separated with spaces. 


```
Example: Run surfstore on 2 block servers

```
> go run cmd/SurfstoreServerExec/main.go -s block -p 8081 -l 
> go run cmd/SurfstoreServerExec/main.go -s block -p 8082 -l 
> go run cmd/SurfstoreServerExec/main.go -s meta -l localhost:8081 localhost:8082

```



## Scalability
updated March,7,2023
#Overview
Consistent hashing is a distributed hashing technique used to evenly distribute data among multiple nodes in a cluster or network. It is commonly used in distributed caching systems and load balancers to ensure that data is stored and accessed efficiently across the network.

#How It Works
In consistent hashing, each node in the cluster is assigned a hash value, typically obtained by hashing the node's IP address or name. Data is also hashed to generate a key, which is then assigned to a node in the cluster based on its hash value.

When a node is added or removed from the cluster, the hash values of all the nodes and the data keys are recalculated to ensure that the data is distributed evenly among the remaining nodes.

One of the benefits of consistent hashing is that it minimizes the number of keys that need to be reassigned when a node is added or removed, as only the keys that were assigned to the node being added or removed are affected.

#Implementation
There are several ways to implement consistent hashing, but the most common one is the ring-based implementation. In this implementation, nodes are placed on a ring, with each node being responsible for the data between its hash value and the hash value of its clockwise neighbor.

To find which node a key should be assigned to, the key's hash value is computed and then located on the ring. The node responsible for the data range containing the key's hash value is the node to which the key is assigned.

When a node is added or removed from the ring, the data ranges for the affected nodes are recalculated, and the keys are reassigned to the appropriate nodes.


## Fault-tolerent

still working on it
