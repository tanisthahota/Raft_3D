package api

import (
	"encoding/json"
	"net/http"
	"time"
	"raft3d/raft"
	hashiraft "github.com/hashicorp/raft"
)

func PostFilament(w http.ResponseWriter, r *http.Request, node *hashiraft.Raft) {
	var f map[string]string
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := raft.Command{
		Op:   "add_filament",
		Data: raft.MustMarshal(f),
	}
	data, _ := json.Marshal(cmd)
	node.Apply(data, 5*time.Second)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(f)
}

func GetFilaments(w http.ResponseWriter, r *http.Request, fsm *raft.FSM) {
	fsm.Mu.Lock()
	defer fsm.Mu.Unlock()

	var filaments []map[string]string
	for _, f := range fsm.Filaments {
		filaments = append(filaments, f)
	}
	json.NewEncoder(w).Encode(filaments)
}