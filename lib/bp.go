package lib

import (
	"image"
	"image/draw"
)

func EmbedBitplane(bitplane_args []string, cover_img image.Image, secret_img image.Image) (image.Image, error) {
	// Parse bitplans operation input
	bitplane_operations, err := BitplaneArgsToArray(bitplane_args)
	if err != nil {
		return nil, err
	}

	// Create cover image copy (for faster pixel read and write)
	cover_bounds := cover_img.Bounds()
	cover_width, cover_height := cover_bounds.Max.X, cover_bounds.Max.Y
	new_cover_img := image.NewNRGBA(cover_bounds)
	draw.Draw(new_cover_img, cover_bounds, cover_img, cover_bounds.Min, draw.Src)

	// Create secret image copy
	secret_bounds := secret_img.Bounds()
	secret_width, secret_height := secret_bounds.Max.X, secret_bounds.Max.Y
	new_secret_img := image.NewNRGBA(secret_bounds)
	draw.Draw(new_secret_img, secret_bounds, secret_img, secret_bounds.Min, draw.Src)

	// Fit secret within cover
	if secret_width > cover_width {
		secret_width = cover_width
	}
	if secret_height > cover_height {
		secret_height = cover_height
	}

	// Check number of values per pixel in image
	secret_values_per_pixel, err := GetValuesPerPixel(new_secret_img)
	if err != nil {
		return nil, err
	}
	cover_values_per_pixel, err := GetValuesPerPixel(new_cover_img)
	if err != nil {
		return nil, err
	}

	// Iterate through image
	for y := 0; y < secret_height; y++ {
		for x := 0; x < secret_width; x++ {
			secret_index := (y*secret_width + x) * secret_values_per_pixel
			for _, embed_instruction := range bitplane_operations {
				cover_index := (y*cover_width + x) * cover_values_per_pixel
				// Get bit position and colour from instruction
				colour := embed_instruction[0].(int)
				bit_pos := embed_instruction[1].(int)
				// check if secret image should be 1 or 0
				secret_pixel := new_secret_img.Pix[secret_index]
				// Change cover image based on above
				if secret_pixel < 127 {
					mask := ^(1 << bit_pos)
					new_cover_img.Pix[cover_index+colour] &= uint8(mask)
				} else {
					new_cover_img.Pix[cover_index+colour] |= (1 << bit_pos)
				}
			}
		}
	}

	return new_cover_img, nil
}

func ExtractBitplane(bitplane_args []string, input_img image.Image) (image.Image, error) {
	// Parse bitplans operation input
	bitplane_operations, err := BitplaneArgsToArray(bitplane_args)
	if err != nil {
		return nil, err
	}

	// Create cover image copy (for faster pixel read and write)
	bounds := input_img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	new_img := image.NewNRGBA(bounds)
	draw.Draw(new_img, bounds, input_img, bounds.Min, draw.Src)

	// Check number of values per pixel in image
	values_per_pixel, err := GetValuesPerPixel(new_img)
	if err != nil {
		return nil, err
	}

	// Iterate through image
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			index := (y*width + x) * values_per_pixel
			pix := new_img.Pix[index : index+values_per_pixel]
			// Check that all bit planes for given operations are set to 1
			has_all_bits := true
			for _, embed_instruction := range bitplane_operations {
				colour := embed_instruction[0].(int)
				bit_pos := embed_instruction[1].(int)
				if !HasBit(pix[colour], bit_pos) {
					has_all_bits = false
				}
			}
			// Set to 255/0 depending on has_all_bits
			new_col := 255
			if has_all_bits {
				new_col = 0
			}
			for i := 0; i < values_per_pixel; i++ {
				new_img.Pix[index+i] = uint8(new_col)
			}
		}
	}

	return new_img, nil
}
