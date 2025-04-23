package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"raft3d/raft"
	hashiraft "github.com/hashicorp/raft"
)

func PostPrintJob(w http.ResponseWriter, r *http.Request, node *hashiraft.Raft) {
	var job map[string]string
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := raft.Command{
		Op:   "add_print_job",
		Data: raft.MustMarshal(job),
	}
	data, _ := json.Marshal(cmd)
	node.Apply(data, 5*time.Second)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

func GetPrintJobs(w http.ResponseWriter, r *http.Request, fsm *raft.FSM) {
	fsm.Mu.Lock()
	defer fsm.Mu.Unlock()

	var jobs []map[string]string
	for _, job := range fsm.PrintJobs {
		jobs = append(jobs, job)
	}
	json.NewEncoder(w).Encode(jobs)
}

func UpdateJobStatus(w http.ResponseWriter, r *http.Request, node *hashiraft.Raft) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/print_jobs/")
	id = strings.TrimSuffix(id, "/status")

	status := r.URL.Query().Get("status")
	if status == "" {
		http.Error(w, "status query param required", http.StatusBadRequest)
		return
	}

	cmd := raft.Command{
		Op: "update_job_status",
		Data: raft.MustMarshal(map[string]string{
			"id":     id,
			"status": status,
		}),
	}
	data, _ := json.Marshal(cmd)
	node.Apply(data, 5*time.Second)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Updated"))
}