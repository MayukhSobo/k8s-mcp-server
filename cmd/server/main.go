package main

import (
	"fmt"
	"os"

	"github.com/mayukhsarkar/k8s-mcp-server/pkg/api"
	"github.com/spf13/cobra"
)

var (
	port       int
	kubeconfig string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "k8s-mcp-server",
		Short: "Kubernetes MCP Server - A backend system for managing Kubernetes resources and logs",
		Long: `Kubernetes MCP Server provides an interactive and extensible interface for 
managing Kubernetes resources, retrieving and analyzing logs, and formatting logs for export.`,
	}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server",
		Long:  "Start the Kubernetes MCP server to handle requests for Kubernetes operations and log management",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Starting Kubernetes MCP Server on port %d\n", port)
			server := api.NewServer(port, kubeconfig)
			if err := server.Start(); err != nil {
				fmt.Printf("Error starting server: %v\n", err)
				os.Exit(1)
			}
		},
	}

	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the server on")
	serveCmd.Flags().StringVarP(&kubeconfig, "kubeconfig", "k", "", "Path to kubeconfig file (defaults to in-cluster config if empty)")

	rootCmd.AddCommand(serveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
