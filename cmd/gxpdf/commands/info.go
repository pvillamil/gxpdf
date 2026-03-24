package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/coregx/gxpdf"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info FILE",
	Short: "Display PDF metadata and information",
	Long: `Display detailed information about a PDF file.

Shows:
  - File size and page count
  - PDF version
  - Document metadata (title, author, subject, keywords)
  - Creation and modification dates
  - Encryption status and permissions
  - Producer and creator applications

Examples:
  gxpdf info document.pdf
  gxpdf info report.pdf --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func runInfo(_ *cobra.Command, args []string) error {
	filePath := args[0]

	// Get file info.
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	doc, err := gxpdf.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer func() { _ = doc.Close() }()

	info := pdfInfo{
		File:      filePath,
		FileSize:  fileInfo.Size(),
		PageCount: doc.PageCount(),
		Version:   doc.Version(),
		Title:     doc.Title(),
		Author:    doc.Author(),
		Subject:   doc.Subject(),
		Keywords:  doc.Keywords(),
		Creator:   doc.Creator(),
		Producer:  doc.Producer(),
		Encrypted: doc.IsEncrypted(),
	}

	switch outputFormat {
	case formatJSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(info)
	default:
		return outputInfoText(info)
	}
}

type pdfInfo struct {
	File      string `json:"file"`
	FileSize  int64  `json:"file_size"`
	PageCount int    `json:"page_count"`
	Version   string `json:"version"`
	Title     string `json:"title,omitempty"`
	Author    string `json:"author,omitempty"`
	Subject   string `json:"subject,omitempty"`
	Keywords  string `json:"keywords,omitempty"`
	Creator   string `json:"creator,omitempty"`
	Producer  string `json:"producer,omitempty"`
	Encrypted bool   `json:"encrypted"`
}

//nolint:unparam // Returns nil for consistency with other output functions.
func outputInfoText(info pdfInfo) error {
	fmt.Printf("File:       %s\n", info.File)
	fmt.Printf("Size:       %s\n", formatSize(info.FileSize))
	fmt.Printf("Pages:      %d\n", info.PageCount)
	fmt.Printf("Version:    PDF %s\n", info.Version)
	fmt.Printf("Encrypted:  %v\n", info.Encrypted)

	if info.Title != "" {
		fmt.Printf("Title:      %s\n", info.Title)
	}
	if info.Author != "" {
		fmt.Printf("Author:     %s\n", info.Author)
	}
	if info.Subject != "" {
		fmt.Printf("Subject:    %s\n", info.Subject)
	}
	if info.Keywords != "" {
		fmt.Printf("Keywords:   %s\n", info.Keywords)
	}
	if info.Creator != "" {
		fmt.Printf("Creator:    %s\n", info.Creator)
	}
	if info.Producer != "" {
		fmt.Printf("Producer:   %s\n", info.Producer)
	}

	return nil
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
