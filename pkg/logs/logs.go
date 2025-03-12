package logs

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// LogManager handles log operations
type LogManager struct {
	clientset *kubernetes.Clientset
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Pod       string    `json:"pod"`
	Container string    `json:"container"`
	Namespace string    `json:"namespace"`
	LogLevel  string    `json:"level,omitempty"`
}

// LogOptions represents options for retrieving logs
type LogOptions struct {
	Namespace    string
	Pod          string
	Container    string
	SinceTime    *time.Time
	SinceSeconds *int64
	Tail         *int64
	Pattern      string
	LogLevel     string
}

// NewLogManager creates a new LogManager
func NewLogManager(clientset *kubernetes.Clientset) *LogManager {
	return &LogManager{
		clientset: clientset,
	}
}

// GetLogs retrieves logs from a pod
func (lm *LogManager) GetLogs(opts LogOptions) ([]LogEntry, error) {
	podLogOpts := corev1.PodLogOptions{
		Container:    opts.Container,
		SinceTime:    nil,
		SinceSeconds: opts.SinceSeconds,
		TailLines:    opts.Tail,
		Follow:       false,
	}

	if opts.SinceTime != nil {
		sinceTime := metav1.NewTime(*opts.SinceTime)
		podLogOpts.SinceTime = &sinceTime
	}

	req := lm.clientset.CoreV1().Pods(opts.Namespace).GetLogs(opts.Pod, &podLogOpts)
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("error opening log stream: %v", err)
	}
	defer podLogs.Close()

	var logEntries []LogEntry
	reader := bufio.NewReader(podLogs)

	// Compile regex pattern if provided
	var re *regexp.Regexp
	if opts.Pattern != "" {
		re, err = regexp.Compile(opts.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %v", err)
		}
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading logs: %v", err)
		}

		// Parse log entry
		entry := parseLogEntry(line, opts.Pod, opts.Container, opts.Namespace)

		// Filter by pattern if provided
		if re != nil && !re.MatchString(entry.Message) {
			continue
		}

		// Filter by log level if provided
		if opts.LogLevel != "" && !strings.EqualFold(entry.LogLevel, opts.LogLevel) {
			continue
		}

		logEntries = append(logEntries, entry)
	}

	return logEntries, nil
}

// parseLogEntry parses a log line into a structured LogEntry
func parseLogEntry(line, pod, container, namespace string) LogEntry {
	// Default timestamp to now
	timestamp := time.Now()

	// Try to extract timestamp from the log line
	// This is a simple implementation and might need to be adjusted based on actual log format
	timestampRegex := regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`)
	if match := timestampRegex.FindString(line); match != "" {
		if t, err := time.Parse(time.RFC3339, match); err == nil {
			timestamp = t
		}
	}

	// Try to extract log level
	logLevel := ""
	logLevelRegex := regexp.MustCompile(`(?i)\b(info|error|warn|debug|fatal)\b`)
	if match := logLevelRegex.FindString(line); match != "" {
		logLevel = strings.ToUpper(match)
	}

	return LogEntry{
		Timestamp: timestamp,
		Message:   strings.TrimSpace(line),
		Pod:       pod,
		Container: container,
		Namespace: namespace,
		LogLevel:  logLevel,
	}
}

// ExportLogs exports logs in the specified format
func (lm *LogManager) ExportLogs(entries []LogEntry, format string, writer io.Writer) error {
	switch strings.ToLower(format) {
	case "json":
		return json.NewEncoder(writer).Encode(entries)
	case "csv":
		return exportToCSV(entries, writer)
	case "ndjson":
		return exportToNDJSON(entries, writer)
	case "plaintext", "text":
		return exportToPlaintext(entries, writer)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// exportToCSV exports logs to CSV format
func exportToCSV(entries []LogEntry, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	if err := csvWriter.Write([]string{"Timestamp", "Pod", "Container", "Namespace", "Level", "Message"}); err != nil {
		return fmt.Errorf("error writing CSV header: %v", err)
	}

	// Write data
	for _, entry := range entries {
		if err := csvWriter.Write([]string{
			entry.Timestamp.Format(time.RFC3339),
			entry.Pod,
			entry.Container,
			entry.Namespace,
			entry.LogLevel,
			entry.Message,
		}); err != nil {
			return fmt.Errorf("error writing CSV record: %v", err)
		}
	}

	return nil
}

// exportToNDJSON exports logs to NDJSON format
func exportToNDJSON(entries []LogEntry, writer io.Writer) error {
	for _, entry := range entries {
		if err := json.NewEncoder(writer).Encode(entry); err != nil {
			return fmt.Errorf("error writing NDJSON record: %v", err)
		}
	}
	return nil
}

// exportToPlaintext exports logs to plaintext format
func exportToPlaintext(entries []LogEntry, writer io.Writer) error {
	for _, entry := range entries {
		line := fmt.Sprintf("[%s] [%s] [%s/%s] [%s] %s\n",
			entry.Timestamp.Format(time.RFC3339),
			entry.Namespace,
			entry.Pod,
			entry.Container,
			entry.LogLevel,
			entry.Message)
		if _, err := writer.Write([]byte(line)); err != nil {
			return fmt.Errorf("error writing plaintext record: %v", err)
		}
	}
	return nil
}
