package main

import (
	"log"
	"net/http"
	"os"
	"raft3d/api"
	"raft3d/raft"
	hashiraft "github.com/hashicorp/raft"
)

func main() {
	nodeID := os.Args[1]
	raftAddr := os.Args[2]
	dataDir := os.Args[3]
	httpPort := os.Args[4]

	r, fsm, err := raft.SetupRaft(nodeID, raftAddr, dataDir)
	if err != nil {
		log.Fatalf("Raft setup failed: %v", err)
	}

	if nodeID == "node1" {
		config := hashiraft.Configuration{
			Servers: []hashiraft.Server{
				{ID: hashiraft.ServerID("node1"), Address: hashiraft.ServerAddress("127.0.0.1:12000")},
				{ID: hashiraft.ServerID("node2"), Address: hashiraft.ServerAddress("127.0.0.1:12001")},
				{ID: hashiraft.ServerID("node3"), Address: hashiraft.ServerAddress("127.0.0.1:12002")},
			},
		}
		r.BootstrapCluster(config)
	}

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

	log.Println("HTTP Server on :" + httpPort)
	log.Fatal(http.ListenAndServe(":"+httpPort, nil))
}
