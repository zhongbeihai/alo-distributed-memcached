# alo-distributed-mencached

# 1. Flow Review

```text
1  Receive key  -->  Check if cached  -------- Yes --------->  Return cached value (1)
2                                  |
3                                  No
4                                  |---->  Should retrieve from remote node  ---- Yes --->  Interact with remote node  -->  Return cached value (2)
5                                  |                                      |
6                                  |                                      No
7                                  |---->  Call callback function, fetch value and add to cache  -->  Return cached value (3)

Detailed for (2)
1  Use consistent hashing to select node
2       |---->  Is it a remote node?  ---- Yes --->  HTTP client requests remote node  ---> Success? ----->  Server returns value
3       |                                        |
4       |                                        No
5       |---->  Return to local node for processing
```

# Project Structure
- **pkg/consistent_hash/**: Implements consistent hashing, which is used to distribute keys across cache nodes efficiently and with minimal rebalancing when nodes join or leave.
- **pkg/lru/**: Contains the LRU (Least Recently Used) cache logic for managing the local in-memory cache.
- **pkg/single_flight/**: Provides a mechanism to ensure that only one request for a given key is in-flight at a time, preventing cache breakdown under high concurrency.
- **pkg/alo_cache.go**: The core logic for the distributed cache, including the Group abstraction, cache lookup, peer selection, and data loading logic.
- **pkg/http.go**: Handles HTTP server and client logic for inter-node communication, including request routing and peer selection.
- **pkg/peers.go**: Defines the PeerPicker and PeerGetter interfaces, and implements HTTPGetter for fetching data from remote nodes.
- **main.go**: The main entry point of the application. Sets up the cache group, configures the HTTP pool (cluster), and starts the HTTP server.


# Dependency
```
+-------------------+
|      Group        |
+-------------------+
         |
         | 1. Needs to select the node responsible for a key
         v
+-------------------+
|   PeerPicker      |<-----------------------------+
|   (interface)     |                              |
+-------------------+                              |
         |                                         |
         | Implemented by HTTPPool                 |
         v                                         |
+-------------------+                              |
|    HTTPPool       |                              |
+-------------------+                              |
         |                                         |
         | 2. Returns a PeerGetter for the node    |
         v                                         |
+-------------------+                              |
|   PeerGetter      |<--------------------------+  |
|   (interface)     |                           |  |
+-------------------+                           |  |
         |                                      |  |
         | Implemented by HTTPGetter            |  |
         v                                      |  |
+-------------------+                           |  |
|   HTTPGetter      |                           |  |
+-------------------+                           |  |
         |                                      |  |
         | 3. Communicates with remote node via |  |
         |    HTTP to fetch data                |  |
         +--------------------------------------|--+
                                                |
+-------------------+                           |
| ConsistentHashMap |<--------------------------+
+-------------------+
         ^
         |
         | Used by HTTPPool for node selection
```