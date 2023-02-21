# surfstore

## overview

![Alt text](../iShot_2023-02-20_17.51.45.png)

This is a cloud-based file storage service called SurfStore. SurfStore is a networked file storage application that is based on Dropbox, and lets you sync files to and from the “cloud”. Cloud services are implemented, and a client is designed which interacts with the service via gRPC.

Multiple clients can concurrently connect to the SurfStore service to access a common, shared set of files. Clients accessing SurfStore “see” a consistent set of updates to files, but SurfStore does not offer any guarantees about operations across files, meaning that it does not support multi-file transactions (such as atomic move).

## Fundamentals

### Blocks, hashes, and hashlists

A file in SurfStore is broken into an ordered sequence of one or more blocks. Each block is of uniform size (defined by the command line argument), except for the last block in the file, which may be smaller (but must be at least 1 byte large). As an example, assume the block size is 4096 bytes, and consider the following file:

![Alt text](../iShot_2023-02-20_17.55.35.png)

### The Base Directory

A command-line argument specifies a “base directory” for the client. This is the directory that is going to be synchronized with your cloud-based service. Your client will upload files from this base directory to the cloud, and download files (and changes to files) from the cloud into this base directory. Your client should not modify any files outside of this base directory. Note in particular–your client should not download files into the “current” directory, only the base directory specified by that command line argument. 

### Version

Each file/filename is associated with a version, which is a monotonically increasing positive integer. The version is incremented any time the file is created, modified, or deleted. The purpose of the version is so that clients can detect when they have an out-of-date view of the file hierarchy.
Note: if two clients try to modify the same file in the cloud concurently, the one that uploads the change later will recieved a "upload failure" flag.

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
go run cmd/SurfstoreServerExec/main.go -s <service_type> -p <port> -l -d (blockstoreAddr*)

Usage:
-s <service_type>: 
(required) This defines the service provided by this server. It can be “meta”, “block”, or “both” (you don’t need to include the quotation marks).
-p <port>:
(default=8080) Port to accept connections
-l:
Only listen on localhost if included
-d:
Output log statements
(blockStoreAddr*):
BlockStore address (ip:port) the MetaStore should be initialized with. (Note: if service_type = both, then you should also include the address of the server that you’re starting)
```
