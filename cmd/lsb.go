/*
	Copyright © 2022 George Omnet stegogo@georgeom.net
*/
package cmd

import (
	"errors"
	"image/png"
	"io/ioutil"
	"log"
	"os"

	"stegogo/lib"

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
	Long:  "Embed a secret within an image via Least Significant Bit steganography.",
	Args: func(cmd *cobra.Command, args []string) error {
		// Ensure we get at least 1 import
		if len(args) < 1 {
			return errors.New(`bit positions must be given, i.e., "R0", or "B2 R1"`)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, bitplane_args []string) error {
		// Parse input flags
		secret_file_path, _ := cmd.Flags().GetString("secret")
		cover_file_path, _ := cmd.Flags().GetString("cover")
		output_file_path, _ := cmd.Flags().GetString("output")

		// Parse secret
		secret_bits, err := lib.FilepathToBitstream(secret_file_path)
		if err != nil {
			return err
		}

		// Open cover_file
		img, err := lib.OpenImage(cover_file_path)
		if err != nil {
			return err
		}

		// Run embed operation
		edited_img, err := lib.EmbedLsb(bitplane_args, secret_bits, img)
		if err != nil {
			return err
		}

		// Write image to file
		outFile, err := os.Create(output_file_path)
		if err != nil {
			log.Fatal(err)
		}
		defer outFile.Close()
		png.Encode(outFile, edited_img)
		return nil
	},
}

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract data",
	Long:  "Extract secret data from within an image via Least Significant Bit steganography.",
	RunE: func(cmd *cobra.Command, bitplane_args []string) error {
		// Open input file
		input_file_path, _ := cmd.Flags().GetString("input")
		output_file_path, _ := cmd.Flags().GetString("output")
		input_img, err := lib.OpenImage(input_file_path)
		if err != nil {
			return err
		}
		extracted_bits, err := lib.ExtractLsb(bitplane_args, input_img)
		if err != nil {
			return err
		}

		bytes_arr := lib.BitstreamToBytes(extracted_bits)
		ioutil.WriteFile(output_file_path, bytes_arr, 0644)
		return nil
	},
}

func init() {
	// Add commands
	rootCmd.AddCommand(lsbCmd)
	lsbCmd.AddCommand(embedCmd)
	lsbCmd.AddCommand(extractCmd)

	// Add persistent flags
	embedCmd.Flags().StringP("secret", "s", "", "(Required) A file to be embedded in the image.")
	embedCmd.Flags().StringP("cover", "c", "", "(Required) A cover image data embedded within.")
	embedCmd.Flags().StringP("output", "o", "output.png", "(Default 'output.png') Output image path.")
	embedCmd.MarkFlagRequired("secret")
	embedCmd.MarkFlagRequired("cover")

	extractCmd.Flags().StringP("input", "i", "", "(Required) Input file with embedded data inside.")
	extractCmd.Flags().StringP("output", "o", "extracted.dat", "(Default 'output.dat') Output extracted data file.")
	extractCmd.MarkFlagRequired("input")

}
