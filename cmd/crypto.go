package cmd

import (
	"fmt"

	"github.com/dzakwan/ipsec-vpn/pkg/crypto"
	"github.com/spf13/cobra"
)

// cryptoCmd represents the crypto command
var cryptoCmd = &cobra.Command{
	Use:   "crypto",
	Short: "Manage cryptographic settings",
	Long:  `Configure and manage cryptographic settings including post-quantum algorithms.`,
}

var cryptoShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show available cryptographic algorithms",
	Run: func(cmd *cobra.Command, args []string) {
		showPostQuantum, _ := cmd.Flags().GetBool("post-quantum")
		showClassic, _ := cmd.Flags().GetBool("classic")

		// If neither flag is specified, show both
		if !showPostQuantum && !showClassic {
			showPostQuantum = true
			showClassic = true
		}

		if showClassic {
			fmt.Println("Classic encryption algorithms:")
			for _, algo := range crypto.ListClassicAlgorithms() {
				fmt.Printf("- %s: %s\n", algo.Name, algo.Description)
			}
			fmt.Println()
		}

		if showPostQuantum {
			fmt.Println("Post-quantum encryption algorithms:")
			for _, algo := range crypto.ListPostQuantumAlgorithms() {
				fmt.Printf("- %s: %s\n", algo.Name, algo.Description)
			}
		}
	},
}

var cryptoTestCmd = &cobra.Command{
	Use:   "test [algorithm]",
	Short: "Test a cryptographic algorithm",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		algorithm := args[0]
		data, _ := cmd.Flags().GetString("data")
		if data == "" {
			data = "This is a test message for encryption"
		}

		// Test the algorithm
		result, err := crypto.TestAlgorithm(algorithm, []byte(data))
		if err != nil {
			fmt.Printf("Error testing algorithm '%s': %v\n", algorithm, err)
			return
		}

		fmt.Printf("Algorithm: %s\n", algorithm)
		fmt.Printf("Original data: %s\n", data)
		fmt.Printf("Encrypted size: %d bytes\n", len(result.Encrypted))
		fmt.Printf("Decryption successful: %v\n", result.DecryptionSuccessful)
		fmt.Printf("Performance:\n")
		fmt.Printf("  Key generation: %v\n", result.KeyGenTime)
		fmt.Printf("  Encryption: %v\n", result.EncryptTime)
		fmt.Printf("  Decryption: %v\n", result.DecryptTime)
	},
}

var cryptoSetDefaultCmd = &cobra.Command{
	Use:   "set-default [algorithm]",
	Short: "Set the default encryption algorithm",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		algorithm := args[0]
		postQuantum, _ := cmd.Flags().GetBool("post-quantum")

		err := crypto.SetDefaultAlgorithm(algorithm, postQuantum)
		if err != nil {
			fmt.Printf("Error setting default algorithm: %v\n", err)
			return
		}

		fmt.Printf("Default %s algorithm set to: %s\n", 
			postQuantumLabel(postQuantum), algorithm)
	},
}

// Helper function to get the label for post-quantum status
func postQuantumLabel(isPostQuantum bool) string {
	if isPostQuantum {
		return "post-quantum"
	}
	return "classic"
}

func init() {
	// Add subcommands to crypto command
	cryptoCmd.AddCommand(cryptoShowCmd)
	cryptoCmd.AddCommand(cryptoTestCmd)
	cryptoCmd.AddCommand(cryptoSetDefaultCmd)

	// Flags for show command
	cryptoShowCmd.Flags().Bool("post-quantum", false, "Show post-quantum algorithms only")
	cryptoShowCmd.Flags().Bool("classic", false, "Show classic algorithms only")

	// Flags for test command
	cryptoTestCmd.Flags().String("data", "", "Data to use for testing encryption")

	// Flags for set-default command
	cryptoSetDefaultCmd.Flags().Bool("post-quantum", false, "Set as default post-quantum algorithm")
}