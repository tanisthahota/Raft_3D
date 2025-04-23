package api

import (
	"encoding/json"
	"net/http"
	"time"
	"raft3d/raft"
	hashiraft "github.com/hashicorp/raft"
)

func PostPrinter(w http.ResponseWriter, r *http.Request, node *hashiraft.Raft) {
	var p map[string]string
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := raft.Command{
		Op:   "add_printer",
		Data: raft.MustMarshal(p),
	}
	data, _ := json.Marshal(cmd)
	node.Apply(data, 5*time.Second)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func GetPrinters(w http.ResponseWriter, r *http.Request, fsm *raft.FSM) {
	fsm.Mu.Lock()
	defer fsm.Mu.Unlock()

	var list []map[string]string
	for _, p := range fsm.Printers {
		list = append(list, p)
	}
	json.NewEncoder(w).Encode(list)
}