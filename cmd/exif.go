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
			fmt.Println(tag.TagName())
			bytesArr, err := tag.GetRawBytes()
			if err != nil {
				return err
			}
			if len(bytesArr) > 500 {
				bytesArr = bytesArr[:500]
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
		ioutil.WriteFile("out.jpeg", b.Bytes(), 0644)
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
	exifExtractCmd.Flags().StringP("image", "i", "", "(Required) Input image with EXIF data inside.")
	exifEmbedCmd.MarkFlagRequired("image")
	exifEmbedCmd.MarkFlagRequired("secret")
	exifExtractCmd.MarkFlagRequired("image")

}
