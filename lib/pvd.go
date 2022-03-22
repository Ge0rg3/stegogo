package lib

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"math"
	"strconv"
)

func CreateRangeTableArray(range_widths []string) ([][]int, error) {
	/*
		Convert a string of ranges to a quantization range table array.
		i.e., {"8", "8", "16", "32", "64", "128"} -> [(0, 7), (8, 15), (16, 31), (32, 63), (64, 127), (128, 255)]
	*/
	var range_table = make([][]int, len(range_widths))
	start := 0
	for index, range_str := range range_widths {
		// Convert range string to int, i.e., "16" to 16
		range_int, err := strconv.Atoi(range_str)
		if err != nil {
			return range_table, errors.New("invalid range width list given")
		}
		// Add start and end position to range table
		range_table[index] = []int{start, start + range_int - 1}
		start += range_int
	}
	return range_table, nil
}

func checkRangeTable(range_table [][]int, pixel_difference int) (int, int) {
	/*
		Determine the number of embeddable/extractable bits from given
		pixel difference, based off range table values
	*/
	for _, _range := range range_table {
		if _range[0] <= pixel_difference && _range[1] >= pixel_difference {
			embeddable_bits := math.Log2(float64(_range[1] + 1 - _range[0]))
			return _range[0], int(embeddable_bits)
		}
	}
	return 0, 0
}

func EmbedPvd(cover_img image.Image, range_table [][]int, secret_bits string, direction string, zigzag bool) (image.Image, error) {
	// Get image details
	bounds := cover_img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	new_img := image.NewGray(bounds)
	draw.Draw(new_img, bounds, cover_img, bounds.Min, draw.Src)
	// Vars for keeping track of state between pixels
	first_in_pair := true
	previous_index := -1
	secret_position := 0

	// Iterate through pixels in given order, based off: https://gist.github.com/Ge0rg3/282dd5671d755acbf13352a7ae8e2d5e
	start_a, start_b := 0, 0
	step_a, step_b := 1, 1
	end_a, end_b := height, width
	if direction != "row" {
		end_a, end_b = width, height
	}
	for a := start_a; a < end_a; a += step_a {
		// Flip direction of row/col if in zigzag pattern
		if zigzag {
			if a%2 == 1 {
				start_b = end_b - 1
				end_b, step_b = -1, -1
			} else {
				start_b = 0
				step_b = 1
				end_b = width
				if direction != "row" {
					end_b = height
				}
			}
		}
		for b := start_b; b != end_b; b += step_b {
			var index int
			if direction == "row" {
				index = (a * width) + b
			} else {
				index = (b * width) + a
			}
			// Iteration process now complete. PVD logic per pixel now starts
			if first_in_pair {
				// If first in pixel pair, continue to next
				previous_index = index
				first_in_pair = false
			} else {
				// Get difference between current pixel and previous pixel
				pixel_difference := int(new_img.Pix[previous_index]) - int(new_img.Pix[index])
				// Find minimum range and number of embeddable bits using range table
				min_range, bit_count := checkRangeTable(range_table, Abs(pixel_difference))
				// Calculate what data to embed
				var bits_to_embed string
				if secret_position+bit_count <= len(secret_bits) {
					bits_to_embed = secret_bits[secret_position : secret_position+bit_count]
				} else {
					bits_to_embed = secret_bits[secret_position:]
				}
				int_to_embed, _ := strconv.ParseInt(bits_to_embed, 2, 64)
				// Find new difference to put between pixels
				new_pixel_difference := min_range + int(int_to_embed)
				// Change pixel values to fit new difference
				var m float64
				if pixel_difference < 0 {
					m = float64(pixel_difference - (new_pixel_difference * -1))
				} else {
					m = float64(pixel_difference - new_pixel_difference)
				}
				m /= 2

				if pixel_difference%2 == 0 {
					new_img.Pix[previous_index] -= uint8(math.Floor(m))
					new_img.Pix[index] += uint8(math.Ceil(m))
				} else {
					new_img.Pix[previous_index] -= uint8(math.Ceil(m))
					new_img.Pix[index] += uint8(math.Floor(m))
				}
				secret_position += bit_count
				first_in_pair = true
			}
			if secret_position >= len(secret_bits) {
				return new_img, nil
			}

		}
	}
	fmt.Printf("WARNING: Image too small with given secret -- only %d/%d bits embedded.\n", secret_position, len(secret_bits))
	return new_img, nil
}
