package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/your-org/ai-k8s-ops/internal/offline"
	"github.com/your-org/ai-k8s-ops/internal/storage/sqlite"
	"github.com/your-org/ai-k8s-ops/pkg/version"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "version":
		fmt.Printf("AI-K8S-OPS CLI v%s\n", version.Version)
	case "offline":
		if len(os.Args) < 3 {
			printOfflineUsage()
			os.Exit(1)
		}
		handleOffline(os.Args[2])
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf("AI-K8S-OPS CLI v%s\n\n", version.Version)
	fmt.Println("Usage: ai-k8s-ops <command> [subcommand] [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  version    Show version info")
	fmt.Println("  offline    Manage offline packages")
}

func printOfflineUsage() {
	fmt.Println("Usage: ai-k8s-ops offline <subcommand> [flags]")
	fmt.Println()
	fmt.Println("Subcommands:")
	fmt.Println("  export     Export offline package (requires network)")
	fmt.Println("  import     Import offline package")
	fmt.Println("  list       List offline packages")
	fmt.Println("  inspect    Inspect offline package file")
}

func handleOffline(subcmd string) {
	switch subcmd {
	case "export":
		handleExport()
	case "import":
		handleImport()
	case "list":
		handleList()
	case "inspect":
		handleInspect()
	default:
		fmt.Printf("Unknown offline subcommand: %s\n", subcmd)
		printOfflineUsage()
		os.Exit(1)
	}
}

func handleExport() {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	name := fs.String("name", "", "Package name (required)")
	osFlag := fs.String("os", "ubuntu,centos", "Target OS list (comma-separated)")
	modules := fs.String("modules", "core,network", "Modules to include (comma-separated)")
	output := fs.String("output", "data/offline", "Output directory")
	dbPath := fs.String("db", "data/ai-k8s-ops.db", "Database path")
	fs.Parse(os.Args[3:])

	if *name == "" {
		fmt.Println("Error: --name is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	osList := strings.Split(*osFlag, ",")
	moduleList := strings.Split(*modules, ",")

	if !offline.ValidateModules(moduleList) {
		fmt.Printf("Error: invalid module. Valid modules: %s\n", strings.Join(offline.ValidModules(), ", "))
		os.Exit(1)
	}
	if !offline.HasRequiredModules(moduleList) {
		fmt.Println("Error: core and network modules are required")
		os.Exit(1)
	}
	if !offline.ValidateOSList(osList) {
		fmt.Printf("Error: invalid OS. Valid OS: %s\n", strings.Join(offline.ValidOSList(), ", "))
		os.Exit(1)
	}

	db, err := sqlite.Init(*dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	packageDB := offline.NewPackageDB(db)
	exporter := offline.NewExporter(packageDB, *output)

	modulesJSON, _ := json.Marshal(moduleList)
	osListJSON, _ := json.Marshal(osList)

	pkg := &offline.OfflinePackage{
		ID:        fmt.Sprintf("cli-%d", os.Getpid()),
		Name:      *name,
		Version:   offline.K8sVersion,
		OSList:    string(osListJSON),
		Modules:   string(modulesJSON),
		Status:    "pending",
		CreatedBy: "cli",
	}

	if err := packageDB.Create(pkg); err != nil {
		fmt.Printf("Error creating package record: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Exporting offline package: %s\n", *name)
	fmt.Printf("  Version: %s\n", offline.K8sVersion)
	fmt.Printf("  OS: %s\n", *osFlag)
	fmt.Printf("  Modules: %s\n", *modules)
	fmt.Printf("  Output: %s\n", *output)
	fmt.Println()

	// Run export synchronously for CLI
	exporter.Export(pkg)

	fmt.Println("Export started. Check status with: ai-k8s-ops offline list")
}

func handleImport() {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	file := fs.String("file", "", "Package file path (required)")
	dbPath := fs.String("db", "data/ai-k8s-ops.db", "Database path")
	fs.Parse(os.Args[3:])

	if *file == "" {
		fmt.Println("Error: --file is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	db, err := sqlite.Init(*dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	packageDB := offline.NewPackageDB(db)
	importer := offline.NewImporter(packageDB, "data/offline")

	fmt.Printf("Importing offline package: %s\n", *file)

	pkg, err := importer.Import(*file, "cli")
	if err != nil {
		fmt.Printf("Import failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Import successful!\n")
	fmt.Printf("  ID: %s\n", pkg.ID)
	fmt.Printf("  Name: %s\n", pkg.Name)
	fmt.Printf("  Version: %s\n", pkg.Version)
	fmt.Printf("  Checksum: %s\n", pkg.Checksum)
}

func handleList() {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	dbPath := fs.String("db", "data/ai-k8s-ops.db", "Database path")
	fs.Parse(os.Args[3:])

	db, err := sqlite.Init(*dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	packageDB := offline.NewPackageDB(db)
	packages, err := packageDB.List()
	if err != nil {
		fmt.Printf("Error listing packages: %v\n", err)
		os.Exit(1)
	}

	if len(packages) == 0 {
		fmt.Println("No offline packages found.")
		return
	}

	fmt.Printf("%-36s %-20s %-10s %-12s %-10s\n", "ID", "NAME", "VERSION", "STATUS", "SIZE")
	fmt.Println(strings.Repeat("-", 90))
	for _, p := range packages {
		size := formatSize(p.Size)
		fmt.Printf("%-36s %-20s %-10s %-12s %-10s\n", p.ID, p.Name, p.Version, p.Status, size)
	}
}

func handleInspect() {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	file := fs.String("file", "", "Package file path (required)")
	fs.Parse(os.Args[3:])

	if *file == "" {
		fmt.Println("Error: --file is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	info, err := os.Stat(*file)
	if err != nil {
		fmt.Printf("Error: file not found: %s\n", *file)
		os.Exit(1)
	}

	fmt.Printf("File: %s\n", *file)
	fmt.Printf("Size: %s\n", formatSize(info.Size()))
	fmt.Println("(Full inspection requires extracting manifest - not yet implemented)")
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
