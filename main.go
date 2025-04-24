package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"raft3d/api"
	"raft3d/raft"
	"time" // Import time package

	hashiraft "github.com/hashicorp/raft"
)

func main() {
	nodeID := os.Args[1]
	raftAddr := os.Args[2]
	dataDir := os.Args[3]
	httpPort := os.Args[4]

	log.Printf("Starting node %s with Raft address %s, data directory %s, HTTP port %s", nodeID, raftAddr, dataDir, httpPort)


	r, fsm, err := raft.SetupRaft(nodeID, raftAddr, dataDir)
	if err != nil {
		log.Fatalf("Raft setup failed: %v", err)
	}

	log.Printf("Node %s initialized. Waiting for cluster formation...", nodeID)

	// Bootstrap logic for node1
	if nodeID == "node1" {
		// Check if the cluster has already been bootstrapped by looking for existing peers.
		currentConfigFuture := r.GetConfiguration()
		if err := currentConfigFuture.Error(); err != nil {
			log.Fatalf("Failed to get current raft configuration: %v", err)
		}
		currentConfig := currentConfigFuture.Configuration()

		// Only bootstrap if there are no servers in the configuration
		if len(currentConfig.Servers) == 0 {
			log.Printf("Node %s attempting to bootstrap cluster.", nodeID)
			bootstrapConfig := hashiraft.Configuration{
				Servers: []hashiraft.Server{
					{ID: hashiraft.ServerID("node1"), Address: hashiraft.ServerAddress("192.168.118.243:12000")},
					{ID: hashiraft.ServerID("node2"), Address: hashiraft.ServerAddress("192.168.118.114:12000")},
					{ID: hashiraft.ServerID("node3"), Address: hashiraft.ServerAddress("192.168.118.200:12000")},
				},
			}
			future := r.BootstrapCluster(bootstrapConfig)
			if err := future.Error(); err != nil {
				// Log error but don't necessarily exit, maybe it joined an existing cluster
				log.Printf("BootstrapCluster error on node %s: %v. This might be okay if cluster already exists.", nodeID, err)
			} else {
				log.Printf("Node %s successfully bootstrapped cluster.", nodeID)
			}
		} else {
			log.Printf("Node %s found existing configuration with %d servers. Skipping bootstrap.", nodeID, len(currentConfig.Servers))
		}
	} else {
		log.Printf("Node %s started as follower/candidate. Waiting to join cluster.", nodeID)
		// Followers should eventually discover the leader via the transport layer
		// if node1 successfully bootstrapped and advertised the configuration.
	}

	// Add a loop to periodically log the leader and peers
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				leaderAddr, leaderID := r.LeaderWithID()
				state := r.State().String() // Get state as string
				term := r.CurrentTerm()
				// Get peers from configuration
				configFuture := r.GetConfiguration()
				peersStr := "unknown"
				if configFuture.Error() == nil {
					peersStr = fmt.Sprintf("%v", configFuture.Configuration().Servers)
				}

				log.Printf("Node %s [State: %s, Term: %d, LeaderID: %s, LeaderAddr: %s, Peers: %s]",
					nodeID, state, term, leaderID, leaderAddr, peersStr)
			// Add a way to stop this goroutine if needed, e.g. context cancellation
			}
		}
	}()


	http.HandleFunc("/api/v1/printers", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			api.PostPrinter(w, req, r)
		} else {
			api.GetPrinters(w, req, fsm)
		}
	})
	http.HandleFunc("/api/v1/filaments", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			api.PostFilament(w, req, r)
		} else {
			api.GetFilaments(w, req, fsm)
		}
	})
	http.HandleFunc("/api/v1/print_jobs", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			api.PostPrintJob(w, req, r)
		} else {
			api.GetPrintJobs(w, req, fsm)
		}
	})
	http.HandleFunc("/api/v1/print_jobs/", func(w http.ResponseWriter, req *http.Request) {
		api.UpdateJobStatus(w, req, r)
	})

	log.Println("HTTP Server starting on :" + httpPort)
	log.Fatal(http.ListenAndServe(":"+httpPort, nil))
}

// Need to import fmt for the logging goroutine
