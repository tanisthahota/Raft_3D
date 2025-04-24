package raft

import (
	"fmt"
	hashiraft "github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"net"
	"os"
	"time"
)

func SetupRaft(nodeID, bindAddr, dataDir string) (*hashiraft.Raft, *FSM, error) {
    config := hashiraft.DefaultConfig()
    config.LocalID = hashiraft.ServerID(nodeID)
    
    // Add these configurations
    config.HeartbeatTimeout = 1000 * time.Millisecond
    config.ElectionTimeout = 1000 * time.Millisecond
    config.LeaderLeaseTimeout = 500 * time.Millisecond
    config.CommitTimeout = 50 * time.Millisecond

    // Set the bind address for the local node
    addr, err := net.ResolveTCPAddr("tcp", bindAddr)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to resolve TCP addr: %v", err)
    }

    // Important: Set the address that this node will use to communicate
    // config.Address = hashiraft.ServerAddress(bindAddr) // REMOVE THIS LINE
    
    // Log the binding address for debugging
    fmt.Printf("Setting up Raft node %s at address %s\n", nodeID, bindAddr)

    transport, err := hashiraft.NewTCPTransport(bindAddr, addr, 3, 10*time.Second, os.Stderr)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to create TCP transport: %v", err)
    }

	logStore, err := raftboltdb.NewBoltStore(fmt.Sprintf("%s/log.bolt", dataDir))
	if err != nil {
		// Add cleanup
		if logStore != nil {
			logStore.Close()
		}
		return nil, nil, fmt.Errorf("failed to create log store: %v", err)
	}

	stableStore, err := raftboltdb.NewBoltStore(fmt.Sprintf("%s/stable.bolt", dataDir))
	if err != nil {
		// Add cleanup
		if logStore != nil {
			logStore.Close()
		}
		if stableStore != nil {
			stableStore.Close()
		}
		return nil, nil, fmt.Errorf("failed to create stable store: %v", err)
	}

	snapshotStore, err := hashiraft.NewFileSnapshotStore(dataDir, 2, os.Stderr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create snapshot store: %v", err)
	}

	fsm := NewFSM()

	node, err := hashiraft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create raft node: %v", err)
	}

	return node, fsm, nil
}