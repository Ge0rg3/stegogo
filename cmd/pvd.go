/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"image/png"
	"log"
	"os"
	"stegogo/lib"
	"strings"

	"github.com/spf13/cobra"
)

// pvdCmd represents the pvd command
var pvdCmd = &cobra.Command{
	Use:   "pvd",
	Short: "Pixel Value Differencing",
	Long: `Pixel Value Differencing (PVD) steganography is a technique proposed
by Da-Chun Wu and Wen-Hsiang Tsai in a 2003 academic paper. Secret data is embedded
by manipulating the difference between pairs of pixels.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var pvdEmbedCmd = &cobra.Command{
	Use:   "embed",
	Short: "Embed data",
	Long:  ".",
	RunE: func(cmd *cobra.Command, bitplane_args []string) error {
		// Parse input flags
		direction, _ := cmd.Flags().GetString("direction")
		zigzag, _ := cmd.Flags().GetBool("zigzag")
		secret_file_path, _ := cmd.Flags().GetString("secret")
		cover_file_path, _ := cmd.Flags().GetString("cover")
		output_file_path, _ := cmd.Flags().GetString("output")

		// Parse secret
		secret_bitstream, err := lib.FilepathToBitstream(secret_file_path)
		secret_bytesarr := make([]byte, len(secret_bitstream))
		for idx, val := range secret_bitstream {
			if val {
				secret_bytesarr[idx] = '1'
			} else {
				secret_bytesarr[idx] = '0'
			}
		}
		secret_bitstring := string(secret_bytesarr)
		if err != nil {
			return err
		}

		// Open cover_file
		img, err := lib.OpenImage(cover_file_path)
		if err != nil {
			return err
		}

		// Create range table
		range_table, err := lib.CreateRangeTableArray(strings.Split("8 8 16 32 64 128", " "))
		if err != nil {
			return err
		}

		// Run embed function
		new_img, err := lib.EmbedPvd(img, range_table, secret_bitstring, direction, zigzag)
		if err != nil {
			return err
		}

		// Write image to file
		outFile, err := os.Create(output_file_path)
		if err != nil {
			log.Fatal(err)
		}
		defer outFile.Close()
		png.Encode(outFile, new_img)
		return nil
	},
}

func init() {
	// Add commands
	rootCmd.AddCommand(pvdCmd)
	pvdCmd.AddCommand(pvdEmbedCmd)

	// Add flags
	pvdCmd.PersistentFlags().StringP("direction", "d", "row", "(Default 'row') Which direction to iterate through the image. Either 'row' or 'column'.")
	pvdCmd.PersistentFlags().BoolP("zigzag", "z", true, "(Default true) Whether to 'zigzag' across rows/cols.")

	pvdEmbedCmd.Flags().StringP("secret", "s", "", "(Required) A file to be embedded in the image.")
	pvdEmbedCmd.Flags().StringP("cover", "c", "", "(Required) A cover image data embedded within.")
	pvdEmbedCmd.Flags().StringP("output", "o", "output.png", "(Default 'output.png') Output image path.")
	pvdEmbedCmd.MarkFlagRequired("secret")
	pvdEmbedCmd.MarkFlagRequired("cover")
}
