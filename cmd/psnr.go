package cmd

import (
	"fmt"
	"stegogo/lib"

	"github.com/bamiaux/rez"
	"github.com/spf13/cobra"
)

// psnrCmd represents the psnr command
var psnrCmd = &cobra.Command{
	Use:   "psnr",
	Short: "Calculate the PSNR of two images.",
	Long:  `Find the peak signal-to-noise ratio (PSNR) between two images. This is useful for quantifying image distortion.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse input flags
		input_file_path, _ := cmd.Flags().GetString("input")
		changed_file_path, _ := cmd.Flags().GetString("changed")

		// Open input file
		img, err := lib.OpenImage(input_file_path)
		if err != nil {
			return err
		}

		// Open changed file
		changed_img, err := lib.OpenImage(changed_file_path)
		if err != nil {
			return err
		}

		// Conduct PSNR
		res, err := rez.Psnr(img, changed_img)
		if err != nil {
			return err
		}

		fmt.Println("PSNR: ", res)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(psnrCmd)

	psnrCmd.Flags().StringP("input", "i", "", "(Required) Original image file.")
	psnrCmd.Flags().StringP("changed", "c", "", "(Required) Changed/different image file.")
	psnrCmd.MarkFlagRequired("input")
	psnrCmd.MarkFlagRequired("changed")

}
