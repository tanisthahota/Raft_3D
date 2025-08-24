# Raft_3D

A distributed 3D printing management system built with Go, implementing the Raft consensus algorithm for high availability and fault tolerance.



## Overview

Raft_3D is a cloud computing project that provides a distributed system for managing 3D printing resources. It uses the Raft consensus algorithm to ensure that all nodes in the cluster maintain a consistent view of the system state, even in the presence of node failures.

The system manages:
- 3D Printers
- Filaments (with tracking of remaining material)
- Print Jobs (with status tracking)


## Architecture

- **Consensus Protocol**: Hashicorp's Raft implementation
- **Storage**: BoltDB (embedded key-value store)
- **API**: RESTful HTTP endpoints
- **Language**: Go
- **Network Distribution**: Can be deployed across multiple physical devices on a network

## Features

- **Distributed Consensus**: All data is replicated across multiple nodes
- **Fault Tolerance**: System continues to operate even if some nodes fail
- **Resource Management**: Track printers, filaments, and print jobs
- **Automatic Filament Tracking**: Remaining filament weight is automatically updated when print jobs complete
- **Network Distribution**: The system can be validated either on different terminals of the same device or on different physical devices over a network, demonstrating true distributed consensus across separate machines.



## How It Works
1. The system uses the Raft consensus algorithm to maintain a replicated log of commands across all nodes.
2. When a client sends a request to any node, the request is forwarded to the leader node.
3. The leader appends the command to its log and replicates it to the other nodes.
4. Once a majority of nodes have confirmed receipt of the log entry, the leader commits it.
5. Each node applies the committed command to its state machine (FSM).
6. The FSM updates its in-memory state and periodically creates snapshots for persistence.

## Technical Details
### Raft Implementation
The system uses Hashicorp's Raft implementation, which provides:

- Leader election
- Log replication
- Safety guarantees
### State Machine
The Finite State Machine (FSM) maintains the system state:

- Printers
- Filaments
- Print Jobs
### Storage
The system uses BoltDB for persistent storage:

- Log store: Records of all commands
- Stable store: Raft metadata
- Snapshot store: Point-in-time snapshots of the system state


## Architecture

- **Consensus Protocol**: Hashicorp's Raft implementation
- **Storage**: BoltDB (embedded key-value store)
- **API**: RESTful HTTP endpoints
- **Language**: Go

## Features

- **Distributed Consensus**: All data is replicated across multiple nodes
- **Fault Tolerance**: System continues to operate even if some nodes fail
- **Resource Management**: Track printers, filaments, and print jobs
- **Automatic Filament Tracking**: Remaining filament weight is automatically updated when print jobs complete

## Getting Started

### Prerequisites

- Go 1.19 or higher
- Multiple terminal windows for running nodes

### Running the Cluster

Start three nodes on separate terminals:

```bash
# Terminal 1
go run main.go node1 127.0.0.1:12000 data1 8000

# Terminal 2
go run main.go node2 127.0.0.1:12001 data2 8001

# Terminal 3
go run main.go node3 127.0.0.1:12002 data3 8002
```
Each command takes the following arguments:

1. Node ID
2. Raft address for inter-node communication
3. Data directory for persistent storage
4. HTTP port for the API
## Using the API
### Add a Printer
``` curl -X POST -H "Content-Type: application/json" -d "{\"id\":\"printer1\",\"company\":\"Creality\",\"model\":\"Ender 3\"}" http://127.0.0.1:8000/api/v1/printers```

### Add Filament
```curl -X POST -H "Content-Type: application/json" -d "{\"id\":\"filament1\",\"type\":\"PLA\",\"color\":\"Blue\",\"total_weight_in_grams\":\"1000\",\"remaining_weight_in_grams\":\"1000\"}" http://127.0.0.1:8000/api/v1/filaments ```

### Add a Print Job
```curl -X POST -H "Content-Type: application/json" -d "{\"id\":\"job1\",\"printer_id\":\"printer1\",\"filament_id\":\"filament1\",\"file_path\":\"/path/to/file.gcode\",\"status\":\"pending\"}" ### Add a Print Job```

### Update Job Status 
``` curl -X POST "http://127.0.0.1:8000/api/v1/print_jobs/job1/status?status=running"```

### Available statuses:

- ```running ```- Marks job as running
- ```done``` - Marks job as complete and updates filament weight
- ```canceled``` - Marks job as canceled

### Get Resources
- Get all filaments
```curl http://127.0.0.1:8000/api/v1/filaments```

-  Get all print jobs
```curl http://127.0.0.1:8000/api/v1/print_jobs```

- Get all printers
```curl http://127.0.0.1:8000/api/v1/printers ```

## Project Improvements

### Performance
- **Optimized Raft Configuration:**  
  Fine-tuned Raft parameters in `raft_node.go`:  
  - HeartbeatTimeout: `1000ms`  
  - ElectionTimeout: `1000ms`  
  - LeaderLeaseTimeout: `500ms`  
  - CommitTimeout: `50ms`  
  → Improves leader election stability and reduces consensus latency.

### Reliability
- **Enhanced Error Handling:**  
  Ensures BoltDB stores are properly closed if errors occur during `SetupRaft`.  
- **Improved Status Handling:**  
  Job status update logic now accepts multiple status formats (case-insensitive) and adds debug logging for state changes.

### Monitoring
- **Cluster State Monitoring:**  
  Added a background goroutine that periodically logs:  
  - Current node state  
  - Term number  
  - Leader info  
  - Peer configuration  
  → Improves operational visibility.

### Deployment
- **Smarter Cluster Bootstrap:**  
  `main.go` checks if the cluster is already configured before bootstrapping to avoid conflicts.  
- **Network Distribution:**  
  Validated across multiple physical devices via IP-based bootstrap settings, proving real distributed consensus.

### API
- **More Flexible API:**  
  Job status endpoint now accepts variations like `running/Running`, `done/Done/completed/Completed`, etc., making it more user-friendly.

## Troubleshooting

- Cannot bind port: Verify no other process is using the Raft or HTTP port.

- Cluster not electing a leader: Check nodes can reach each other (firewall, IP/port correctness).

- State not persisting: Confirm the data directory is writable and not ephemeral.
