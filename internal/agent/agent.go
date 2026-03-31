package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// Config holds agent configuration
type Config struct {
	ServerURL  string
	ClusterID  string
	Token      string
	Interval   time.Duration
}

// Agent represents a cluster agent
type Agent struct {
	config  Config
	client  *http.Client
}

// NodeStatus represents node status report
type NodeStatus struct {
	ClusterID    string            `json:"cluster_id"`
	NodeName     string            `json:"node_name"`
	Role         string            `json:"role"`
	Status       string            `json:"status"`
	CPUUsage     float64           `json:"cpu_usage"`
	MemoryUsage  float64           `json:"memory_usage"`
	DiskUsage    float64           `json:"disk_usage"`
	PodCount     int               `json:"pod_count"`
	Labels       map[string]string `json:"labels"`
	LastHeartbeat time.Time        `json:"last_heartbeat"`
}

// CommandRequest represents a command execution request
type CommandRequest struct {
	ID        string   `json:"id"`
	ClusterID string   `json:"cluster_id"`
	Command   []string `json:"command"`
	Timeout   int      `json:"timeout"`
}

// CommandResponse represents a command execution response
type CommandResponse struct {
	ID         string `json:"id"`
	Success    bool   `json:"success"`
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"`
	ExitCode   int    `json:"exit_code"`
}

// NewAgent creates a new agent instance
func NewAgent(config Config) *Agent {
	return &Agent{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Run starts the agent main loop
func (a *Agent) Run(ctx context.Context) error {
	log.Println("Agent starting...")

	// Register with server
	if err := a.register(ctx); err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}
	log.Println("Agent registered successfully")

	ticker := time.NewTicker(a.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := a.heartbeat(ctx); err != nil {
				log.Printf("Heartbeat failed: %v", err)
			}
			if err := a.checkCommands(ctx); err != nil {
				log.Printf("Command check failed: %v", err)
			}
		}
	}
}

// register registers the agent with the server
func (a *Agent) register(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/agents/register", a.config.ServerURL)

	payload := map[string]string{
		"cluster_id": a.config.ClusterID,
		"hostname":   getHostname(),
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
	}

	return a.postJSON(ctx, url, payload, nil)
}

// heartbeat sends heartbeat to server
func (a *Agent) heartbeat(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/agents/heartbeat", a.config.ServerURL)

	status, err := a.collectStatus()
	if err != nil {
		return fmt.Errorf("status collection failed: %w", err)
	}

	return a.postJSON(ctx, url, status, nil)
}

// checkCommands checks for pending commands
func (a *Agent) checkCommands(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/agents/commands?cluster_id=%s", a.config.ServerURL, a.config.ClusterID)

	var commands []CommandRequest
	if err := a.getJSON(ctx, url, &commands); err != nil {
		return err
	}

	for _, cmd := range commands {
		go a.executeCommand(cmd)
	}

	return nil
}

// executeCommand executes a remote command
func (a *Agent) executeCommand(req CommandRequest) {
	log.Printf("Executing command: %v", req.Command)

	ctx := context.Background()
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.Timeout)*time.Second)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, req.Command[0], req.Command[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	resp := CommandResponse{
		ID:         req.ID,
	}

	err := cmd.Run()
	if err != nil {
		resp.Success = false
		resp.Error = err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			resp.ExitCode = exitErr.ExitCode()
		} else {
			resp.ExitCode = -1
		}
	} else {
		resp.Success = true
		resp.Output = stdout.String()
		resp.ExitCode = 0
	}

	// Report result
	url := fmt.Sprintf("%s/api/v1/agents/commands/result", a.config.ServerURL)
	a.postJSON(ctx, url, resp, nil)
}

// collectStatus collects current node status
func (a *Agent) collectStatus() (*NodeStatus, error) {
	status := &NodeStatus{
		ClusterID:     a.config.ClusterID,
		NodeName:      getHostname(),
		Role:          "worker", // TODO: detect from k8s
		Status:        "Ready",
		LastHeartbeat: time.Now(),
		Labels:        make(map[string]string),
	}

	// Collect CPU usage
	if cpu, err := getCPUUsage(); err == nil {
		status.CPUUsage = cpu
	}

	// Collect memory usage
	if mem, err := getMemoryUsage(); err == nil {
		status.MemoryUsage = mem
	}

	// Collect disk usage
	if disk, err := getDiskUsage(); err == nil {
		status.DiskUsage = disk
	}

	// Collect pod count (if kubectl available)
	if pods, err := getPodCount(); err == nil {
		status.PodCount = pods
	}

	return status, nil
}

// HTTP helpers

func (a *Agent) postJSON(ctx context.Context, url string, payload interface{}, result interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.config.Token)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed: %s - %s", resp.Status, string(body))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (a *Agent) getJSON(ctx context.Context, url string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+a.config.Token)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed: %s - %s", resp.Status, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// System helpers

func getHostname() string {
	if name, err := os.Hostname(); err == nil {
		return name
	}
	return "unknown"
}

func getCPUUsage() (float64, error) {
	// Simple implementation - can be improved with /proc/stat
	return 0.0, nil
}

func getMemoryUsage() (float64, error) {
	// Simple implementation - can be improved with /proc/meminfo
	return 0.0, nil
}

func getDiskUsage() (float64, error) {
	cmd := exec.Command("df", "/", "--output=pcent")
	out, err := cmd.Output()
	if err != nil {
		return 0.0, err
	}
	// Parse output (format: "Use%\n 23%")
	var usage float64
	fmt.Sscanf(string(out), "Use%\n %f%%", &usage)
	return usage, nil
}

func getPodCount() (int, error) {
	cmd := exec.Command("kubectl", "get", "pods", "--all-namespaces", "--no-headers")
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	// Count lines
	count := 0
	for _, b := range out {
		if b == '\n' {
			count++
		}
	}
	return count, nil
}
