package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mjasion/k8s-supabase/scripts/docker-compose-migration/pkg/converter"
	"github.com/mjasion/k8s-supabase/scripts/docker-compose-migration/pkg/fetcher"
	"github.com/mjasion/k8s-supabase/scripts/docker-compose-migration/pkg/helm"
	"github.com/mjasion/k8s-supabase/scripts/docker-compose-migration/pkg/kustomize"
	"github.com/mjasion/k8s-supabase/scripts/docker-compose-migration/pkg/parser"
)

const (
	defaultSupabaseRepo   = "https://github.com/supabase/supabase"
	defaultBranch         = "master"
	defaultDockerPath     = "docker"
	defaultOutputDir      = "../../"
)

func main() {
	// Parse command-line flags
	repoURL := flag.String("repo", defaultSupabaseRepo, "Supabase repository URL")
	branch := flag.String("branch", defaultBranch, "Git branch to fetch from")
	dockerPath := flag.String("docker-path", defaultDockerPath, "Path to docker directory in repo")
	outputDir := flag.String("output", defaultOutputDir, "Output directory for generated files")
	skipFetch := flag.Bool("skip-fetch", false, "Skip fetching from GitHub (use local files)")
	localPath := flag.String("local", "", "Path to local docker-compose directory")
	flag.Parse()

	fmt.Println("🚀 Supabase Docker Compose to Kubernetes Converter")
	fmt.Println("==================================================")

	// Step 1: Fetch or load docker-compose files
	var composeData []byte
	var envData []byte
	var err error

	if *skipFetch && *localPath != "" {
		fmt.Printf("📁 Loading local files from: %s\n", *localPath)
		composeData, err = os.ReadFile(filepath.Join(*localPath, "docker-compose.yml"))
		if err != nil {
			log.Fatalf("Failed to read local docker-compose.yml: %v", err)
		}
		envData, err = os.ReadFile(filepath.Join(*localPath, ".env.example"))
		if err != nil {
			log.Printf("Warning: Failed to read local .env.example: %v", err)
			envData = []byte{}
		}
	} else {
		fmt.Printf("📥 Fetching from GitHub: %s (branch: %s)\n", *repoURL, *branch)
		f := fetcher.NewGitHubFetcher(*repoURL, *branch, *dockerPath)
		composeData, envData, err = f.FetchFiles()
		if err != nil {
			log.Fatalf("Failed to fetch files: %v", err)
		}
	}

	// Step 2: Parse docker-compose configuration
	fmt.Println("🔍 Parsing docker-compose configuration...")
	p := parser.NewParser()
	config, err := p.Parse(composeData, envData)
	if err != nil {
		log.Fatalf("Failed to parse docker-compose: %v", err)
	}

	fmt.Printf("   Found %d services\n", len(config.Services))
	for name := range config.Services {
		fmt.Printf("   - %s\n", name)
	}

	// Step 3: Convert to Kubernetes manifests
	fmt.Println("🔄 Converting to Kubernetes manifests...")
	c := converter.NewConverter(config)
	k8sResources, err := c.Convert()
	if err != nil {
		log.Fatalf("Failed to convert to Kubernetes: %v", err)
	}

	// Step 4: Generate Kustomize structure
	fmt.Println("📦 Generating Kustomize manifests...")
	kustomizeDir := filepath.Join(*outputDir, "kustomize")
	kg := kustomize.NewGenerator(k8sResources, kustomizeDir)
	if err := kg.Generate(); err != nil {
		log.Fatalf("Failed to generate Kustomize: %v", err)
	}
	fmt.Printf("   ✅ Kustomize manifests written to: %s\n", kustomizeDir)

	// Step 5: Generate Helm chart
	fmt.Println("📊 Generating Helm chart...")
	chartDir := filepath.Join(*outputDir, "charts", "supabase")
	hg := helm.NewGenerator(k8sResources, config, chartDir)
	if err := hg.Generate(); err != nil {
		log.Fatalf("Failed to generate Helm chart: %v", err)
	}
	fmt.Printf("   ✅ Helm chart written to: %s\n", chartDir)

	fmt.Println("\n✨ Conversion completed successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review generated manifests in kustomize/ and charts/")
	fmt.Println("  2. Update secret values in values.yaml")
	fmt.Println("  3. Test deployment: helm install supabase charts/supabase")
	fmt.Println("  4. Or with Kustomize: kubectl apply -k kustomize/base")
}
