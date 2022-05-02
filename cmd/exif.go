package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	exif "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	jpeg "github.com/dsoprea/go-jpeg-image-structure/v2"
	"github.com/spf13/cobra"
)

// exifCmd represents the exif command
var exifCmd = &cobra.Command{
	Use:   "exif",
	Short: "EXIF Data Manipulation",
	Long:  `Read, write and edit EXIF blocks in an image.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

var exifExtractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract/read data",
	RunE: func(cmd *cobra.Command, args []string) error {
		input_file_path, _ := cmd.Flags().GetString("image")
		input_tag, _ := cmd.Flags().GetString("tag")
		output_file_path, _ := cmd.Flags().GetString("output")
		rawExif, err := exif.SearchFileAndExtractExif(input_file_path)
		if err != nil {
			return err
		}

		im, err := exifcommon.NewIfdMappingWithStandard()
		if err != nil {
			return err
		}

		ti := exif.NewTagIndex()
		_, index, err := exif.Collect(im, ti, rawExif)
		if err != nil {
			return err
		}

		exifTags := index.RootIfd.DumpTags()
		for _, tag := range exifTags {
			if tag.IsThumbnailOffset() {
				continue
			}
			bytesArr, err := tag.GetRawBytes()
			if err != nil {
				return err
			}
			if input_tag == "" {
				// If input tag not given, display tag to user
				if len(bytesArr) > 100 {
					bytesArr = bytesArr[:100]
				}
				fmt.Printf("**********\n%s:\n%s\n", tag.TagName(), string(bytesArr))
			} else {
				// Otherwise, dump tag to file
				if input_tag == tag.TagName() {
					ioutil.WriteFile(output_file_path, bytesArr, 0644)
				}
			}
		}
		return nil
	},
}

var exifEmbedCmd = &cobra.Command{
	Use:   "embed",
	Short: "Embed EXIF field",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse inputs
		input_image_path, _ := cmd.Flags().GetString("image")
		output_file_path, _ := cmd.Flags().GetString("output")
		exif_tag, _ := cmd.Flags().GetString("exiftag")
		secret_data_path, _ := cmd.Flags().GetString("secret")

		// Read secret
		secret_file, err := os.Open(secret_data_path)
		if err != nil {
			return err
		}
		defer secret_file.Close()
		stats, err := secret_file.Stat()
		if err != nil {
			return err
		}

		var size int64 = stats.Size()
		secret_bytes := make([]byte, size)
		bufr := bufio.NewReader(secret_file)
		bufr.Read(secret_bytes)

		// Get current exif
		rawExif, err := exif.SearchFileAndExtractExif(input_image_path)
		if err != nil {
			return err
		}
		im, err := exifcommon.NewIfdMappingWithStandard()
		if err != nil {
			return err
		}

		ti := exif.NewTagIndex()
		_, index, err := exif.Collect(im, ti, rawExif)
		if err != nil {
			return err
		}

		ifdPath := "IFD0"
		ib := exif.NewIfdBuilderFromExistingChain(index.RootIfd)
		childIb, err := exif.GetOrCreateIbFromRootIb(ib, ifdPath)
		if err != nil {
			return err
		}

		err = childIb.SetStandardWithName(exif_tag, secret_bytes)
		if err != nil {
			return err
		}

		// Write changes
		jmp := jpeg.NewJpegMediaParser()
		intfc, _ := jmp.ParseFile(input_image_path)
		sl := intfc.(*jpeg.SegmentList)
		sl.SetExif(ib)
		b := new(bytes.Buffer)
		sl.Write(b)
		ioutil.WriteFile(output_file_path, b.Bytes(), 0644)
		return nil

	},
}

func init() {
	rootCmd.AddCommand(exifCmd)
	exifCmd.AddCommand(exifExtractCmd)
	exifCmd.AddCommand(exifEmbedCmd)

	exifEmbedCmd.Flags().StringP("image", "i", "", "(Required) Input image with EXIF data inside.")
	exifEmbedCmd.Flags().StringP("exiftag", "e", "ProcessingSoftware", "(Optional) New EXIF tag name.")
	exifEmbedCmd.Flags().StringP("secret", "s", "", "(Required) Secret data filename to be embedded within EXIF tag.")
	exifEmbedCmd.Flags().StringP("output", "o", "output.jpg", "(Default 'output.jpg') Output image path.")
	exifEmbedCmd.MarkFlagRequired("image")
	exifEmbedCmd.MarkFlagRequired("secret")

	exifExtractCmd.Flags().StringP("image", "i", "", "(Required) Input image with EXIF data inside.")
	exifExtractCmd.Flags().StringP("tag", "t", "", "(Optional) Exif tag name to save raw data from. Otherwise, snippers from all tags will be shown.")
	exifExtractCmd.Flags().StringP("output", "o", "output.dat", "(Default 'output.dat') Output file name for tag dump.")
	exifExtractCmd.MarkFlagRequired("image")

}
