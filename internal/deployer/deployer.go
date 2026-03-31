package deployer

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ClusterConfig holds cluster configuration
type ClusterConfig struct {
	Name            string
	KubernetesVersion string
	PodNetworkCIDR  string
	ServiceCIDR     string
	APIServerPort   int
}

// DefaultConfig returns default cluster configuration
func DefaultConfig() *ClusterConfig {
	return &ClusterConfig{
		Name:              "ai-k8s-cluster",
		KubernetesVersion: "v1.28.0",
		PodNetworkCIDR:    "192.168.0.0/16",
		ServiceCIDR:       "10.96.0.0/12",
		APIServerPort:     6443,
	}
}

// DeploymentStatus represents deployment progress
type DeploymentStatus struct {
	Phase       string `json:"phase"`       // initializing, installing, configuring, ready, failed
	Step        string `json:"step"`        // Current step description
	Progress    int    `json:"progress"`    // 0-100
	Error       string `json:"error,omitempty"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// Deployer handles K8s cluster deployment
type Deployer struct {
	config  *ClusterConfig
	status  *DeploymentStatus
	workDir string
}

// NewDeployer creates a new deployer instance
func NewDeployer(config *ClusterConfig) *Deployer {
	if config == nil {
		config = DefaultConfig()
	}
	return &Deployer{
		config: config,
		status: &DeploymentStatus{
			Phase:     "pending",
			StartedAt: time.Now(),
		},
	}
}

// Deploy runs the full deployment process
func (d *Deployer) Deploy(ctx context.Context) error {
	log.Printf("Starting K8s cluster deployment: %s", d.config.Name)

	d.updateStatus("initializing", "Preparing deployment environment", 5)

	// Create work directory
	var err error
	d.workDir, err = os.MkdirTemp("", "k8s-deploy-*")
	if err != nil {
		return d.fail(fmt.Errorf("failed to create work dir: %w", err))
	}
	defer os.RemoveAll(d.workDir)

	// Pre-flight checks
	if err := d.preflightChecks(); err != nil {
		return d.fail(fmt.Errorf("preflight checks failed: %w", err))
	}

	d.updateStatus("installing", "Installing container runtime", 15)

	// Install container runtime
	if err := d.installContainerRuntime(); err != nil {
		return d.fail(fmt.Errorf("container runtime installation failed: %w", err))
	}

	d.updateStatus("installing", "Installing Kubernetes binaries", 30)

	// Install K8s binaries (if not present)
	if err := d.installK8sBinaries(); err != nil {
		return d.fail(fmt.Errorf("K8s binaries installation failed: %w", err))
	}

	d.updateStatus("installing", "Initializing Kubernetes cluster", 50)

	// Initialize cluster
	if err := d.initCluster(); err != nil {
		return d.fail(fmt.Errorf("cluster initialization failed: %w", err))
	}

	d.updateStatus("configuring", "Installing network plugin (Calico)", 70)

	// Install network plugin
	if err := d.installNetworkPlugin(); err != nil {
		return d.fail(fmt.Errorf("network plugin installation failed: %w", err))
	}

	d.updateStatus("configuring", "Configuring kubectl", 85)

	// Setup kubeconfig
	if err := d.setupKubeconfig(); err != nil {
		return d.fail(fmt.Errorf("kubeconfig setup failed: %w", err))
	}

	d.updateStatus("ready", "Deployment completed successfully", 100)
	d.status.CompletedAt = time.Now()

	log.Printf("Cluster deployment completed: %s", d.config.Name)
	return nil
}

// preflightChecks runs pre-deployment checks
func (d *Deployer) preflightChecks() error {
	log.Println("Running preflight checks...")

	// Check if running as root
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root")
	}

	// Check disk space
	if err := checkDiskSpace(5 * 1024); err != nil { // 5GB
		return err
	}

	// Check required binaries
	required := []string{"docker", "curl"}
	for _, bin := range required {
		if _, err := exec.LookPath(bin); err != nil {
			return fmt.Errorf("required binary not found: %s", bin)
		}
	}

	// Check if swap is enabled
	if isSwapEnabled() {
		log.Println("Warning: Swap is enabled, disabling...")
		exec.Command("swapoff", "-a").Run()
	}

	return nil
}

// installContainerRuntime installs and configures containerd
func (d *Deployer) installContainerRuntime() error {
	log.Println("Installing container runtime (containerd)...")

	// Check if docker is already installed
	if _, err := exec.LookPath("docker"); err == nil {
		log.Println("Docker already installed, skipping container runtime installation")
		return nil
	}

	// Install containerd
	cmd := exec.Command("bash", "-c", `
		apt-get update && apt-get install -y containerd
		mkdir -p /etc/containerd
		containerd config default > /etc/containerd/config.toml
		sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
		systemctl restart containerd
	`)
	return runCommand(cmd, os.Stdout)
}

// installK8sBinaries installs kubeadm, kubelet, kubectl
func (d *Deployer) installK8sBinaries() error {
	// Check if already installed
	if _, err := exec.LookPath("kubeadm"); err == nil {
		log.Println("Kubernetes binaries already installed")
		return nil
	}

	log.Println("Installing Kubernetes binaries...")

	cmd := exec.Command("bash", "-c", fmt.Sprintf(`
		apt-get update && apt-get install -y apt-transport-https ca-certificates curl
		mkdir -p /etc/apt/keyrings
		curl -fsSL https://pkgs.k8s.io/core:/stable:/%s/deb/Release.key | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
		echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/%s/deb/ /" > /etc/apt/sources.list.d/kubernetes.list
		apt-get update
		apt-get install -y kubelet kubeadm kubectl
		apt-mark hold kubelet kubeadm kubectl
	`, d.config.KubernetesVersion, d.config.KubernetesVersion))

	return runCommand(cmd, os.Stdout)
}

// initCluster initializes the K8s cluster with kubeadm
func (d *Deployer) initCluster() error {
	log.Println("Initializing Kubernetes cluster...")

	// Generate kubeadm config
	configPath := filepath.Join(d.workDir, "kubeadm-config.yaml")
	config := fmt.Sprintf(`
apiVersion: kubeadm.k8s.io/v1beta3
kind: InitConfiguration
localAPIEndpoint:
  advertiseAddress: "0.0.0.0"
  bindPort: %d
nodeRegistration:
  criSocket: unix:///var/run/containerd/containerd.sock
---
apiVersion: kubeadm.k8s.io/v1beta3
kind: ClusterConfiguration
kubernetesVersion: %s
networking:
  podSubnet: %s
  serviceSubnet: %s
`, d.config.APIServerPort, d.config.KubernetesVersion, d.config.PodNetworkCIDR, d.config.ServiceCIDR)

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return err
	}

	// Run kubeadm init
	cmd := exec.Command("kubeadm", "init", "--config", configPath, "--skip-token-print")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubeadm init failed: %w\n%s", err, string(output))
	}

	log.Println("Cluster initialized successfully")
	return nil
}

// installNetworkPlugin installs Calico CNI
func (d *Deployer) installNetworkPlugin() error {
	log.Println("Installing Calico network plugin...")

	// Download and apply Calico manifest
	cmd := exec.Command("kubectl", "create", "-f", "https://raw.githubusercontent.com/projectcalico/calico/v3.26.1/manifests/calico.yaml")
	return runCommand(cmd, os.Stdout)
}

// setupKubeconfig configures kubectl for the user
func (d *Deployer) setupKubeconfig() error {
	log.Println("Setting up kubeconfig...")

	// Create .kube directory
	homeDir, _ := os.UserHomeDir()
	kubeDir := filepath.Join(homeDir, ".kube")
	if err := os.MkdirAll(kubeDir, 0755); err != nil {
		return err
	}

	// Copy admin.conf
	cmd := exec.Command("cp", "-f", "/etc/kubernetes/admin.conf", filepath.Join(kubeDir, "config"))
	if err := cmd.Run(); err != nil {
		return err
	}

	// Fix permissions
	os.Chmod(filepath.Join(kubeDir, "config"), 0600)

	log.Println("Kubeconfig configured at ~/.kube/config")
	return nil
}

// GetJoinCommand returns the join command for worker nodes
func (d *Deployer) GetJoinCommand() (string, error) {
	cmd := exec.Command("kubeadm", "token", "create", "--print-join-command")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// Status returns current deployment status
func (d *Deployer) Status() *DeploymentStatus {
	return d.status
}

func (d *Deployer) updateStatus(phase, step string, progress int) {
	d.status.Phase = phase
	d.status.Step = step
	d.status.Progress = progress
	log.Printf("[%s] %s (%d%%)", phase, step, progress)
}

func (d *Deployer) fail(err error) error {
	d.status.Phase = "failed"
	d.status.Error = err.Error()
	d.status.CompletedAt = time.Now()
	log.Printf("Deployment failed: %v", err)
	return err
}

// Helper functions

func checkDiskSpace(requiredMB int) error {
	cmd := exec.Command("df", "-m", "/")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	// Parse output (format: Filesystem 1M-blocks Used Available Use% Mounted)
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("unexpected df output")
	}

	var available int
	fmt.Sscanf(lines[1], "%*s %*d %*d %d", &available)

	if available < requiredMB {
		return fmt.Errorf("insufficient disk space: %dMB available, %dMB required", available, requiredMB)
	}
	return nil
}

func isSwapEnabled() bool {
	cmd := exec.Command("swapon", "--show")
	output, _ := cmd.Output()
	return len(output) > 0
}

func runCommand(cmd *exec.Cmd, output io.Writer) error {
	cmd.Stdout = output
	cmd.Stderr = output
	return cmd.Run()
}

func runCommandWithOutput(cmd *exec.Cmd) (string, error) {
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// DeployOfflinePackage deploys from an offline package
func DeployOfflinePackage(ctx context.Context, packagePath string) error {
	log.Printf("Deploying from offline package: %s", packagePath)

	// Extract package
	workDir, err := os.MkdirTemp("", "k8s-offline-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workDir)

	cmd := exec.Command("tar", "-xzf", packagePath, "-C", workDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract package: %w", err)
	}

	// Run install script
	installScript := filepath.Join(workDir, "scripts", "install.sh")
	cmd = exec.Command("bash", installScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
