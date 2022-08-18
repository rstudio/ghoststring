package ghoststring_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	top = func() string {
		topBytes, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err != nil {
			panic(err)
		}

		return strings.TrimSpace(string(topBytes))
	}()
)

func getEphemeralAddr(t *testing.T) string {
	l, err := net.Listen("tcp4", "")
	defer func() { _ = l.Close() }()

	require.Nil(t, err)

	return l.Addr().String()
}

func startIntegrationServer(ctx context.Context, t *testing.T, l *log.Logger, name string, env []string) *exec.Cmd {
	buf := &bytes.Buffer{}

	proc := exec.CommandContext(ctx, filepath.Join(top, "build", runtime.GOOS, runtime.GOARCH, name))
	proc.Env = env
	proc.Stdout = buf
	proc.Stderr = buf

	err := proc.Start()

	if !assert.Nil(t, err) {
		l.Println("buffer:\n", buf.String())
		panic(err.Error())
	}

	return proc
}

func killWaitProc(l *log.Logger, port string, proc *exec.Cmd) {
	if proc.Process == nil {
		l.Println("no process to kill")
		return
	}

	l.Println("port=", port, " buffer:\n", proc.Stdout.(*bytes.Buffer).String())

	req, err := http.NewRequest(http.MethodDelete, "http://127.0.0.1:"+port, nil)
	if err != nil {
		l.Printf("failed to build stop signal request: %[1]v", err)
		return
	}

	l.Printf("sending stop signal to process %[1]v", proc.Process)

	if _, err := http.DefaultClient.Do(req); err != nil {
		l.Printf("failed to make stop signal request: %[1]v", err)
		return
	}

	l.Printf("killing process %[1]v", proc.Process)

	if err := proc.Process.Kill(); err != nil {
		l.Printf("failed to kill process: %[1]v", err)
		return
	}

	l.Printf("waiting for command %[1]v", proc)

	if err := proc.Wait(); err != nil {
		l.Printf("failed to wait for process: %[1]v", err)
	}
}

func waitForHealthy(ctx context.Context, l *log.Logger, port string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:"+port+"/healthcheck", nil)
		if err != nil {
			l.Printf("port=%[1]v request error: %[2]v", port, err)
			return err
		}

		resp, err := http.DefaultClient.Do(req)
		if resp != nil && err == nil && resp.StatusCode == http.StatusOK {
			l.Printf("port=%[1]v healthy!", port)
			return nil
		}

		if err != nil {
			l.Printf("port=%[1]v response error: %[2]v", port, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			l.Printf("port=%[1]v response status: %[2]v", port, resp.StatusCode)
			continue
		}
	}
}

type naiveShape struct {
	Name       string   `json:"name"`
	Complaints []string `json:"complaints"`
}

type naiveMyth struct {
	Name   string       `json:"name"`
	Rating int64        `json:"rating"`
	Shapes []naiveShape `json:"shapes"`
}

func TestIntegrationViaClientSession(t *testing.T) {
	if os.Getenv("GHOSTSTRING_INTEGRATION_TESTING") != "on" {
		t.SkipNow()
	}

	r := require.New(t)

	localTmp := filepath.Join(top, ".local", "tmp")
	r.Nil(os.MkdirAll(localTmp, 0755))

	lf, err := os.Create(filepath.Join(localTmp, fmt.Sprintf("testlog.%[1]v", time.Now().Unix())))
	r.Nil(err)
	defer func() { _ = lf.Close() }()

	l := log.New(lf, "", log.LstdFlags)

	l.Println("starting integration suite")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	secretKeyBytes := make([]byte, 41)
	_, err = rand.Read(secretKeyBytes)
	r.Nil(err)

	rectAddr := getEphemeralAddr(t)
	mythAddr := getEphemeralAddr(t)

	sharedEnv := append(
		os.Environ(),
		[]string{
			"GHOSTSTRING_INTEGRATION_TEST_SECRET_KEY=" + string(secretKeyBytes),
			"GHOSTSTRING_INTEGRATION_TEST_RECTANGLES_ADDR=" + rectAddr,
			"GHOSTSTRING_INTEGRATION_TEST_MYTHS_ADDR=" + mythAddr,
		}...,
	)

	_, rectPort, err := net.SplitHostPort(rectAddr)
	r.Nil(err)

	_, mythPort, err := net.SplitHostPort(mythAddr)
	r.Nil(err)

	rectProc := startIntegrationServer(ctx, t, l, "rectangles", sharedEnv)
	r.Nil(waitForHealthy(ctx, l, rectPort))

	mythProc := startIntegrationServer(ctx, t, l, "myths", sharedEnv)
	r.Nil(waitForHealthy(ctx, l, mythPort))

	rectURLString := "http://127.0.0.1:" + rectPort

	l.Printf("GET %[1]q", rectURLString)

	rectResp, err := http.Get(rectURLString)
	r.Nil(err)

	defer rectResp.Body.Close()

	rectBody := map[string]naiveShape{}
	r.Nil(json.NewDecoder(rectResp.Body).Decode(&rectBody))

	r.Len(rectBody, 3)

	nm := naiveMyth{
		Name:   "birds",
		Rating: 42,
		Shapes: []naiveShape{},
	}

	for _, ns := range rectBody {
		nm.Shapes = append(
			nm.Shapes,
			ns,
		)
	}

	reqBody := map[string]any{
		"threshold":  42,
		"collection": []naiveMyth{nm},
	}

	reqBytes, err := json.Marshal(reqBody)
	r.Nil(err)

	mythResp, err := http.Post("http://127.0.0.1:"+mythPort, "application/json", bytes.NewReader(reqBytes))
	r.Nil(err)
	r.NotNil(mythResp)

	mythRespBody := map[string]int64{"score": 0}
	r.Nil(json.NewDecoder(mythResp.Body).Decode(&mythRespBody))

	r.Equal(int64(1218), mythRespBody["score"])

	l.Println("done with assertions")

	killWaitProc(l, rectPort, rectProc)
	killWaitProc(l, mythPort, mythProc)
}
