# CS4032_Project
CS4032 Project: Distributed File System

The four features that I implemented are:
Distributed Transparent File Access
Directory Service
Caching
Lock Service

The directory server stores per file directory listings and also acts as a lock service.
As well as that it stores a  timestamp for each file in order to handle caching.
If a client already has the latest version of the file cached it does not fetch it again in order to save bandwidth.
File servers register with the directory service during initialization. Servers are assigned to files on a round-robin basis.
The filesystem follows the AFS model: when a file is opened it is downloaded from the file server if necessary, when the file is closed it is written back to the file server.
Reads and writes are done on the locally cached copy.
When opening a file a lock must first be acquired otherwise the file will not be opened and an error message will be returned to the client.
After closing the file the lock is released.
To store files the clients and servers use boltdb, an embedded database which stores data in a memory mapped file on the hard disk and delegates in-memory caching to the OS.
This means that commonly used files and other data are cached in RAM for faster access.
The client proxy provides a transparent file access interface to the distributed filesystem, similar to the standard library file implementation.
An example client which uses the client proxy library is included in this repository.
Compile and start bash scripts are provided for the servers and client.
The servers can be configured by modifying the config file or providing command line arguments to the start script.

