package offline

const K8sVersion = "v1.28.0"

func GetResourceManifest() *ResourceManifest {
	return &ResourceManifest{
		Version: K8sVersion,
		Modules: []ModuleInfo{
			{
				Name:        "core",
				Required:    true,
				Description: "K8S 核心组件",
				Images: []string{
					"registry.k8s.io/kube-apiserver:v1.28.0",
					"registry.k8s.io/kube-controller-manager:v1.28.0",
					"registry.k8s.io/kube-scheduler:v1.28.0",
					"registry.k8s.io/kube-proxy:v1.28.0",
					"registry.k8s.io/etcd:3.5.9-0",
					"registry.k8s.io/coredns/coredns:v1.10.1",
					"registry.k8s.io/pause:3.9",
				},
				Binaries: []string{
					"kubeadm",
					"kubelet",
					"kubectl",
					"crictl",
					"containerd",
				},
				EstimatedSize: "2.1GB",
			},
			{
				Name:        "network",
				Required:    true,
				Description: "网络插件 (Calico)",
				Images: []string{
					"docker.io/calico/cni:v3.26.1",
					"docker.io/calico/node:v3.26.1",
					"docker.io/calico/kube-controllers:v3.26.1",
				},
				EstimatedSize: "320MB",
			},
			{
				Name:        "monitoring",
				Required:    false,
				Description: "Prometheus + Grafana",
				Images: []string{
					"quay.io/prometheus/prometheus:v2.47.0",
					"grafana/grafana:10.1.0",
					"quay.io/prometheus/node-exporter:v1.6.1",
					"registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.10.0",
				},
				EstimatedSize: "800MB",
			},
			{
				Name:        "logging",
				Required:    false,
				Description: "Loki 日志系统",
				Images: []string{
					"grafana/loki:2.9.1",
					"grafana/promtail:2.9.1",
				},
				EstimatedSize: "200MB",
			},
			{
				Name:        "tracing",
				Required:    false,
				Description: "Jaeger 追踪系统",
				Images: []string{
					"jaegertracing/all-in-one:1.49",
				},
				EstimatedSize: "150MB",
			},
		},
	}
}

// ValidModules returns list of valid module names
func ValidModules() []string {
	return []string{"core", "network", "monitoring", "logging", "tracing"}
}

// ValidOSList returns list of supported OS
func ValidOSList() []string {
	return []string{"ubuntu", "centos"}
}

// ValidateModules checks if all module names are valid
func ValidateModules(modules []string) bool {
	valid := make(map[string]bool)
	for _, m := range ValidModules() {
		valid[m] = true
	}
	for _, m := range modules {
		if !valid[m] {
			return false
		}
	}
	return true
}

// ValidateOSList checks if all OS names are valid
func ValidateOSList(osList []string) bool {
	valid := make(map[string]bool)
	for _, os := range ValidOSList() {
		valid[os] = true
	}
	for _, os := range osList {
		if !valid[os] {
			return false
		}
	}
	return true
}

// HasRequiredModules checks if required modules are present
func HasRequiredModules(modules []string) bool {
	moduleSet := make(map[string]bool)
	for _, m := range modules {
		moduleSet[m] = true
	}
	manifest := GetResourceManifest()
	for _, mod := range manifest.Modules {
		if mod.Required && !moduleSet[mod.Name] {
			return false
		}
	}
	return true
}
