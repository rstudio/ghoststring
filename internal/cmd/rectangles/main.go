package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rstudio/ghoststring"
)

type rectangle struct {
	Name       string                    `json:"name"`
	Ratio      float64                   `json:"ratio"`
	Complaints []ghoststring.GhostString `json:"complaints"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	secretKey := os.Getenv("GHOSTSTRING_INTEGRATION_TEST_SECRET_KEY")
	addr := os.Getenv("GHOSTSTRING_INTEGRATION_TEST_RECTANGLES_ADDR")

	gh, err := ghoststring.NewAES256GCMSingleKeyGhostifyer("hightops", secretKey)
	if err != nil {
		log.Fatal(err)
	}

	if err := ghoststring.SetGhostifyer(gh); err != nil {
		log.Fatal(err)
	}

	db := map[string]rectangle{
		"square": rectangle{
			Name:  "square",
			Ratio: 1.,
			Complaints: []ghoststring.GhostString{
				{
					Namespace: "hightops",
					Str:       "predictable",
				},
				{
					Namespace: "hightops",
					Str:       "inflexible",
				},
			},
		},
		"tallboi": rectangle{
			Name:  "tallboi",
			Ratio: 2.,
			Complaints: []ghoststring.GhostString{
				{
					Namespace: "hightops",
					Str:       "arrogant",
				},
			},
		},
		"wideboi": rectangle{
			Name:  "wideboi",
			Ratio: 0.5,
			Complaints: []ghoststring.GhostString{
				{
					Str: "overly mysterious",
				},
			},
		},
	}

	appFunc := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Server", "rectangles/0")

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

		if req.Method != http.MethodGet {
			http.Error(w, "not like this", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(db); err != nil {
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
