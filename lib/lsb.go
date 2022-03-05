package lib

import (
	"fmt"
	"image"
	"image/draw"
)

func EmbedLsb(bitplane_args []string, secret_bitstream []bool, cover_img image.Image) (image.Image, error) {
	// Parse bitplans operation input
	bitplane_operations, err := BitplaneArgsToArray(bitplane_args)
	if err != nil {
		return nil, err
	}

	// Create image copy (for faster pixel read and write)
	bounds := cover_img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	new_img := image.NewRGBA(bounds)
	draw.Draw(new_img, bounds, cover_img, bounds.Min, draw.Src)

	// Iterate through all pixels and embed data
	secret_pos := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Get RGBA values for pixel
			index := (y*width + x) * 4
			pix := new_img.Pix[index : index+4]
			for _, embed_instruction := range bitplane_operations {
				// Get bit position and colour from instruction
				colour := embed_instruction[0].(int)
				bit_pos := embed_instruction[1].(int)
				// Flip bit to either 0 or 1 depending on secret data
				if secret_bitstream[secret_pos] {
					pix[colour] |= (1 << bit_pos)
				} else {
					mask := ^(1 << bit_pos)
					pix[colour] &= uint8(mask)
				}
				// Return if secret stream end reached
				secret_pos += 1
				if secret_pos == len(secret_bitstream) {
					return new_img, nil
				}
			}
		}
	}
	fmt.Printf("WARNING: Image too small with given bitplane inputs -- only %d/%d bits embedded.\n", secret_pos, len(secret_bitstream))
	return new_img, nil
}

func ExtractLsb(bitplane_args []string, input_img image.Image) ([]bool, error) {
	// Parse bitplans operation input
	bitplane_operations, err := BitplaneArgsToArray(bitplane_args)
	if err != nil {
		return nil, err
	}

	// Open image as readable object
	bounds := input_img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	parsable_img := image.NewRGBA(bounds)
	draw.Draw(parsable_img, bounds, input_img, bounds.Min, draw.Src)

	// bitstream := ""
	var bitstream = make([]bool, height*width*len(bitplane_operations))
	bitstream_pos := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Get RGBA values for pixel
			index := (y*width + x) * 4
			pix := parsable_img.Pix[index : index+4]
			for _, embed_instruction := range bitplane_operations {
				colour := embed_instruction[0].(int)
				bit_pos := embed_instruction[1].(int)
				if HasBit(pix[colour], bit_pos) {
					bitstream[bitstream_pos] = true
				} else {
					bitstream[bitstream_pos] = false
				}
				bitstream_pos += 1
			}
		}
	}
	return bitstream, nil
}
