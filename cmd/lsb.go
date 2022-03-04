/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// lsbCmd represents the lsb command
var lsbCmd = &cobra.Command{
	Use:   "lsb",
	Short: "Least Significant Bit Steganography",
	Long: `Least Significant Bit (LSB) Steganography is a technique where
a secret bitstream is embedded into different bits of each pixel in an image
by flipping the bit to match that of the secret data.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var embedCmd = &cobra.Command{
	Use:   "embed",
	Short: "Embed data",
	Long:  "Embed a secret within an image.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("embed called")
	},
}

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract data",
	Long:  "Extract secret data from within an image.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("extract called")
	},
}

func init() {
	// Add commands
	rootCmd.AddCommand(lsbCmd)
	lsbCmd.AddCommand(embedCmd)
	lsbCmd.AddCommand(extractCmd)

	// Add persistent flags
	embedCmd.PersistentFlags().StringP("secretFile", "sF", "", "A file to be embedded in the image.")
	embedCmd.PersistentFlags().StringP("coverFile", "cF", "", "A cover image data embedded within.")
	embedCmd.MarkPersistentFlagRequired("secretFile")
	embedCmd.MarkPersistentFlagRequired("coverFile")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// lsbCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
