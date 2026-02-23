package commands

import (
	"fmt"

	"github.com/coregx/gxpdf"
	"github.com/spf13/cobra"
)

var (
	decryptPassword string
	decryptOutput   string
)

var decryptCmd = &cobra.Command{
	Use:   "decrypt FILE -p PASSWORD -o OUTPUT",
	Short: "Decrypt password-protected PDF",
	Long: `Decrypt a password-protected PDF file.

Validates that the password can open the encrypted PDF.
Writing a decrypted copy is not yet supported.

Examples:
  gxpdf decrypt encrypted.pdf -p mypassword -o decrypted.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: runDecrypt,
}

func init() {
	decryptCmd.Flags().StringVarP(&decryptPassword, "password", "p", "", "Password to decrypt (required)")
	decryptCmd.Flags().StringVarP(&decryptOutput, "output", "o", "", "Output file (not yet supported)")
	_ = decryptCmd.MarkFlagRequired("password")
}

func runDecrypt(_ *cobra.Command, args []string) error {
	inputFile := args[0]

	// Try to open the encrypted PDF with the given password
	doc, err := gxpdf.OpenWithPassword(inputFile, decryptPassword)
	if err != nil {
		return fmt.Errorf("failed to open encrypted PDF: %w", err)
	}
	defer doc.Close()

	info := doc.Info()
	fmt.Printf("Successfully opened encrypted PDF: %s\n", inputFile)
	fmt.Printf("  Pages: %d\n", info.PageCount)
	fmt.Printf("  Version: %s\n", info.Version)

	if decryptOutput != "" {
		return fmt.Errorf("writing decrypted PDF is not yet supported\n\nThe PDF was successfully opened and validated with the given password")
	}

	return nil
}
