package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/your-org/ai-k8s-ops/internal/deploy"
)

func main() {
	db, err := sql.Open("sqlite3", "data/ai-k8s-ops.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	templateDB := deploy.NewTemplateDB(db)

	templates := []*deploy.DeploymentTemplate{
		{
			ID:          "dev-template",
			Name:        "开发环境模板",
			Description: "单节点 K8S 集群，适合开发测试",
			Type:        "dev",
			Provider:    "bare-metal",
			Config:      `{"nodes": 1, "version": "v1.28.0", "network": "calico"}`,
			Components:  `["prometheus", "grafana"]`,
			IsDefault:   true,
		},
		{
			ID:          "test-template",
			Name:        "测试环境模板",
			Description: "3节点 K8S 集群，包含完整监控栈",
			Type:        "test",
			Provider:    "bare-metal",
			Config:      `{"nodes": 3, "version": "v1.28.0", "network": "calico"}`,
			Components:  `["prometheus", "grafana", "loki", "jaeger"]`,
			IsDefault:   false,
		},
		{
			ID:          "prod-template",
			Name:        "生产环境模板",
			Description: "高可用 K8S 集群，适合生产环境",
			Type:        "prod",
			Provider:    "bare-metal",
			Config:      `{"masters": 3, "workers": 3, "version": "v1.28.0", "network": "calico", "ha": true}`,
			Components:  `["prometheus", "grafana", "loki", "jaeger", "alertmanager"]`,
			IsDefault:   false,
		},
	}

	for _, t := range templates {
		err := templateDB.Create(t)
		if err != nil {
			log.Printf("Failed to create template %s: %v", t.Name, err)
		} else {
			log.Printf("Created template: %s", t.Name)
		}
	}

	log.Println("Seed completed!")
}
