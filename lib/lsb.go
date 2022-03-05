package lib

import (
	"fmt"
	"image"
	"image/draw"
)

func EmbedLsb(bitplane_args []string, secret_bitstream []bool, cover_img image.Image, order string) (image.Image, error) {
	// Parse bitplans operation input
	bitplane_operations, err := BitplaneArgsToArray(bitplane_args)
	if err != nil {
		return nil, err
	}

	// Create image copy (for faster pixel read and write)
	bounds := cover_img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	new_img := image.NewNRGBA(bounds)
	draw.Draw(new_img, bounds, cover_img, bounds.Min, draw.Src)

	// Iterate through all pixels and embed data
	secret_pos := 0
	// Determine whether to iterate via rows or cols
	var xy_1 int
	var xy_2 int
	if order == "row" {
		xy_1, xy_2 = height, width
	} else {
		xy_1, xy_2 = width, height
	}
	for a := 0; a < xy_1; a++ {
		for b := 0; b < xy_2; b++ {
			// Determine pixel value based on whether we are iterating by row or col
			var index int
			if order == "row" {
				index = (a*width + b) * 4
			} else {
				index = (b * width * 4) + (a * 4)
			}
			// Get pixel colours
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

func ExtractLsb(bitplane_args []string, input_img image.Image, order string) ([]bool, error) {
	// Parse bitplans operation input
	bitplane_operations, err := BitplaneArgsToArray(bitplane_args)
	if err != nil {
		return nil, err
	}

	// Open image as readable object
	bounds := input_img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	parsable_img := image.NewNRGBA(bounds)
	draw.Draw(parsable_img, bounds, input_img, bounds.Min, draw.Src)

	// Determine which way to read image
	var xy_1 int
	var xy_2 int
	if order == "row" {
		xy_1, xy_2 = height, width
	} else {
		xy_1, xy_2 = width, height
	}
	var bitstream = make([]bool, height*width*len(bitplane_operations))
	bitstream_pos := 0
	for a := 0; a < xy_1; a++ {
		for b := 0; b < xy_2; b++ {
			// Determine pixel value based on whether we are iterating by row or col
			var index int
			if order == "row" {
				index = (a*width + b) * 4
			} else {
				index = (b * width * 4) + (a * 4)
			}
			// Get RGBA values for pixel
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
