/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"stegogo/lib"

	"github.com/spf13/cobra"
)

// bpCmd represents the bp command
var bpCmd = &cobra.Command{
	Use:   "bp",
	Short: "Bitplane",
	Long: `The Bitplane module is used to extract or emebed images within
a given bitplane within an image. This is done by replacing a given bit per pixel
with either 1 or 0, depending on a given cover image.

For embedding, supply an array of bit plane segments (i.e. "R0", or "R0 B2") and the image will be embedded into all of them.
For extraction, supply an array of bit plane segments and an image will be rendered where *all* bits are set in the given planes.
`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var bpEmbedCmd = &cobra.Command{
	Use:   "embed",
	Short: "Embed data",
	Long:  "Embed a cover image within any given bit plane/bit plane combination.",
	Args: func(cmd *cobra.Command, args []string) error {
		// Ensure we get at least 1 import
		if len(args) < 1 {
			return errors.New(`bit positions must be given, i.e., "R0", or "B2 R1"`)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, bitplane_args []string) error {
		// Parse input flags
		cover_file_path, _ := cmd.Flags().GetString("cover")
		secret_file_path, _ := cmd.Flags().GetString("secret")
		output_file_path, _ := cmd.Flags().GetString("output")

		// Open cover_file
		cover_img, err := lib.OpenImage(cover_file_path)
		if err != nil {
			return err
		}

		// Open secret_file
		secret_img, err := lib.OpenImage(secret_file_path)
		if err != nil {
			return err
		}

		new_img, err := lib.EmbedBitplane(bitplane_args, cover_img, secret_img)
		if err != nil {
			return err
		}

		// Write image to file
		out_file, err := os.Create(output_file_path)
		if err != nil {
			log.Fatal(err)
		}
		defer out_file.Close()
		png.Encode(out_file, new_img)

		return nil
	},
}

var bpExtractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract data",
	Long:  "Extract a given single-channel image from any given bit plane.",
	Args: func(cmd *cobra.Command, args []string) error {
		// Ensure we get at least 1 import
		if len(args) < 1 {
			return errors.New(`bit positions must be given, i.e., "R0", or "B2 R1"`)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, bitplane_args []string) error {
		// Parse input flags
		input_file_path, _ := cmd.Flags().GetString("input")
		output_file_path, _ := cmd.Flags().GetString("output")

		// Open input file
		input_img, err := lib.OpenImage(input_file_path)
		if err != nil {
			return err
		}

		// Extract image from bit planes
		new_img, err := lib.ExtractBitplane(bitplane_args, input_img)
		if err != nil {
			return err
		}

		// Write image to file
		out_file, err := os.Create(output_file_path)
		if err != nil {
			log.Fatal(err)
		}
		defer out_file.Close()
		png.Encode(out_file, new_img)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(bpCmd)
	bpCmd.AddCommand(bpEmbedCmd)
	bpCmd.AddCommand(bpExtractCmd)

	bpEmbedCmd.Flags().StringP("cover", "c", "", "(Required) A cover image file for the secret image to be embedded within.")
	bpEmbedCmd.Flags().StringP("secret", "s", "", "(Required) A one-channel secret image file to be embedded within the cover image.")
	bpEmbedCmd.Flags().StringP("output", "o", "output.png", "(Default 'output.png') Output image path.")
	bpEmbedCmd.MarkFlagRequired("secret")
	bpEmbedCmd.MarkFlagRequired("cover")

	bpExtractCmd.Flags().StringP("input", "i", "", "(Required) Input file with embedded data inside.")
	bpExtractCmd.Flags().StringP("output", "o", "output.png", "Output file for extracted bit plane slice.")
	bpEmbedCmd.MarkFlagRequired("input")
	bpEmbedCmd.MarkFlagRequired("output")

}
