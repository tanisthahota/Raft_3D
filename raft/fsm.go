package raft

import (
	"encoding/json"
	"io"
	"log"
	"sync"

	hashiraft"github.com/hashicorp/raft"
)

type Command struct {
	Op   string          `json:"op"`
	Data json.RawMessage `json:"data"`
}

// FSM implements the Raft FSM interface
type FSM struct {
	Mu        sync.Mutex  // Exported mutex for access from other packages
	Printers  map[string]map[string]string
	Filaments map[string]map[string]string
	PrintJobs map[string]map[string]string
}

func NewFSM() *FSM {
	return &FSM{
		Printers:  make(map[string]map[string]string),
		Filaments: make(map[string]map[string]string),
		PrintJobs: make(map[string]map[string]string),
	}
}

func (f *FSM) Apply(logEntry *hashiraft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(logEntry.Data, &cmd); err != nil {
		log.Printf("Failed to unmarshal log: %v", err)
		return nil
	}

	f.Mu.Lock()
	defer f.Mu.Unlock()

	switch cmd.Op {
	case "add_printer":
		var p map[string]string
		if err := json.Unmarshal(cmd.Data, &p); err == nil {
			f.Printers[p["id"]] = p
		}
	case "add_filament":
		var fData map[string]string
		if err := json.Unmarshal(cmd.Data, &fData); err == nil {
			f.Filaments[fData["id"]] = fData
		}
	case "add_print_job":
		var job map[string]string
		if err := json.Unmarshal(cmd.Data, &job); err == nil {
			job["status"] = "Queued"
			f.PrintJobs[job["id"]] = job
		}
	case "update_job_status":
		var update map[string]string
		if err := json.Unmarshal(cmd.Data, &update); err == nil {
			id := update["id"]
			status := update["status"]
			log.Printf("Updating job %s to status %s", id, status)  // Add debug log
			
			if job, exists := f.PrintJobs[id]; exists {
				log.Printf("Current job status: %s", job["status"])  // Add debug log
				
				switch status {
				case "running", "Running":
					if job["status"] == "Queued" {
						job["status"] = "Running"
						log.Printf("Updated to Running")  // Add debug log
					}
				case "done", "Done", "completed", "Completed":
					if job["status"] == "Running" {
						job["status"] = "Done"
						fid := job["filament_id"]
						fil := f.Filaments[fid]
						rem := parseInt(fil["remaining_weight_in_grams"])
						used := parseInt(job["print_weight_in_grams"])
						fil["remaining_weight_in_grams"] = intToString(rem - used)
					}
				case "canceled", "Canceled", "failed", "Failed":
					if job["status"] == "Queued" || job["status"] == "Running" {
						job["status"] = "Canceled"
					}
				}
				// Update the job in the map
				f.PrintJobs[id] = job
				log.Printf("Final job status: %s", f.PrintJobs[id]["status"])  // Add debug log
			} else {
				log.Printf("Job %s not found", id)  // Add debug log
			}
		}
	}
	return nil
}

func (f *FSM) Snapshot() (hashiraft.FSMSnapshot, error) {
	// In the Snapshot method:
	f.Mu.Lock()
	defer f.Mu.Unlock()

	state := map[string]map[string]map[string]string{
		"printers":  f.Printers,
		"filaments": f.Filaments,
		"printjobs": f.PrintJobs,
	}
	buf, err := json.Marshal(state)
	if err != nil {
		return nil, err
	}
	return &fsmSnapshot{state: buf}, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	state := make(map[string]map[string]map[string]string)
	if err := json.NewDecoder(rc).Decode(&state); err != nil {
		return err
	}
	// In the Restore method:
	f.Mu.Lock()
	defer f.Mu.Unlock()
	f.Printers = state["printers"]
	f.Filaments = state["filaments"]
	f.PrintJobs = state["printjobs"]
	return nil
}

type fsmSnapshot struct {
	state []byte
}

func (s *fsmSnapshot) Persist(sink hashiraft.SnapshotSink) error {
	if _, err := sink.Write(s.state); err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *fsmSnapshot) Release() {}

func parseInt(s string) int {
	var i int
	json.Unmarshal([]byte(s), &i)
	return i
}

func intToString(i int) string {
	b, _ := json.Marshal(i)
	return string(b)
}