package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rstudio/ghoststring"
)

type mythEvaluationRequest struct {
	Threshold  int64  `json:"threshold"`
	Collection []myth `json:"collection"`
}

type myth struct {
	Name   string  `json:"name"`
	Rating int64   `json:"rating"`
	Shapes []shape `json:"shapes"`
}

type shape struct {
	Name       string                    `json:"name"`
	Complaints []ghoststring.GhostString `json:"complaints"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	secretKey := os.Getenv("GHOSTSTRING_INTEGRATION_TEST_SECRET_KEY")
	addr := os.Getenv("GHOSTSTRING_INTEGRATION_TEST_MYTHS_ADDR")

	if _, err := ghoststring.SetGhostifyer("hightops", secretKey); err != nil {
		log.Fatal(err)
	}

	appFunc := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Server", "myths/0")

		if req.URL.Path == "/healthcheck" && req.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			return
		}

		if req.URL.Path != "/" {
			http.Error(w, "not here", http.StatusNotFound)
			return
		}

		if req.Method == http.MethodDelete {
			stop()
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if req.Method != http.MethodPost {
			http.Error(w, "not like this", http.StatusMethodNotAllowed)
			return
		}

		defer req.Body.Close()

		reqBody := &mythEvaluationRequest{Collection: []myth{}}
		if err := json.NewDecoder(req.Body).Decode(reqBody); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":%[1]q}`, err.Error()), http.StatusBadRequest)
			return
		}

		log.Printf("scoring request: %+#[1]v", reqBody)

		score := int64(0)

		for _, m := range reqBody.Collection {
			for _, s := range m.Shapes {
				for _, c := range s.Complaints {
					score += m.Rating * int64(len(c.String))
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")

		if score < reqBody.Threshold {
			w.WriteHeader(http.StatusTeapot)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if err := json.NewEncoder(w).Encode(map[string]int64{"score": score}); err != nil {
			log.Printf("OH NO: %[1]v", err)
		}
	})

	go func() {
		defer stop()

		log.Printf("listening at %[1]q", addr)

		if err := http.ListenAndServe(addr, appFunc); err != nil {
			log.Printf("OH NO: %[1]v", err)
		}
	}()

	<-ctx.Done()
}
