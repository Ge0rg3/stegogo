package lib

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"math"
	"strconv"
	"strings"
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
			err_msg := fmt.Sprintf("invalid range width item '%s' given", range_str)
			return range_table, errors.New(err_msg)
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

func EmbedPvd(cover_img image.Image, range_table [][]int, secret_bits string, direction string, zigzag bool, plane string) (image.Image, error) {
	/*
		Embed a binstring ("11001011") into an image from a given range table, in
		either "row" or "column" direction, optionally in zigzag pattern.
	*/
	// Get R/G/B/A plane if given
	rgba_index, err := RgbaToInt(plane)
	if err != nil {
		return nil, err
	}

	// Get image details and create new type based off given input
	bounds := cover_img.Bounds()
	gray_img := image.NewGray(bounds)
	rgba_img := image.NewRGBA(bounds)
	width, height := bounds.Max.X, bounds.Max.Y
	draw.Draw(gray_img, bounds, cover_img, bounds.Min, draw.Src)
	draw.Draw(rgba_img, bounds, cover_img, bounds.Min, draw.Src)

	// Check number of values per pixel in image
	values_per_pixel, err := GetValuesPerPixel(cover_img)
	if err != nil {
		return nil, err
	}
	var pix_arr []uint8
	if values_per_pixel == 1 {
		pix_arr = gray_img.Pix
	} else {
		pix_arr = rgba_img.Pix
	}

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
				index = ((a*width)+b)*values_per_pixel + rgba_index
			} else {
				index = (b * width * values_per_pixel) + a + rgba_index
			}
			// Iteration process now complete. PVD logic per pixel now starts
			if first_in_pair {
				// If first in pixel pair, continue to next
				previous_index = index
				first_in_pair = false
			} else {
				// Get difference between current pixel and previous pixel
				pixel_difference := int(pix_arr[previous_index]) - int(pix_arr[index])
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

				prev_val := int(pix_arr[previous_index])
				curr_val := int(pix_arr[index])
				if pixel_difference%2 == 0 {
					prev_val -= int(math.Floor(m))
					curr_val += int(math.Ceil(m))
				} else {
					prev_val -= int(math.Ceil(m))
					curr_val += int(math.Floor(m))
				}

				// Fix overflowing values
				if prev_val < 0 {
					curr_val += prev_val * -1
					prev_val = 0
				} else if curr_val < 0 {
					prev_val += curr_val * -1
					curr_val = 0
				} else if prev_val > 255 {
					curr_val -= (prev_val - 255)
					prev_val = 255
				} else if curr_val > 255 {
					prev_val -= (curr_val - 255)
					curr_val = 255
				}

				pix_arr[previous_index] = uint8(prev_val)
				pix_arr[index] = uint8(curr_val)

				secret_position += bit_count
				first_in_pair = true
			}
			if secret_position >= len(secret_bits) {
				if values_per_pixel == 1 {
					return gray_img, nil
				} else {
					return rgba_img, nil
				}
			}

		}
	}
	fmt.Printf("WARNING: Image too small with given secret -- only %d/%d bits embedded.\n", secret_position, len(secret_bits))
	if values_per_pixel == 1 {
		return gray_img, nil
	} else {
		return rgba_img, nil
	}
}

func ExtractPvd(img image.Image, range_table [][]int, direction string, zigzag bool, plane string) ([]byte, error) {
	// Get R/G/B/A plane if given
	rgba_index, err := RgbaToInt(plane)
	if err != nil {
		return nil, err
	}

	// Get image details
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	// Check number of values per pixel in image
	var pix_arr []uint8
	values_per_pixel, err := GetValuesPerPixel(img)
	if err != nil {
		return nil, err
	}
	if values_per_pixel == 1 {
		gray_img := image.NewGray(bounds)
		draw.Draw(gray_img, bounds, img, bounds.Min, draw.Src)
		pix_arr = gray_img.Pix
	} else {
		rgba_img := image.NewRGBA(bounds)
		draw.Draw(rgba_img, bounds, img, bounds.Min, draw.Src)
		pix_arr = rgba_img.Pix
	}

	first_in_pair := true
	previous_index := -1
	var extracted_binstring strings.Builder

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
				index = ((a*width)+b)*values_per_pixel + rgba_index
			} else {
				index = (b * width * values_per_pixel) + a + rgba_index
			}
			// Iteration process now complete. PVD logic per pixel now starts
			if first_in_pair {
				// If first in pixel pair, continue to next
				previous_index = index
				first_in_pair = false
			} else {
				// Get difference between current pixel and previous pixel
				abs_pixel_difference := Abs(int(pix_arr[previous_index]) - int(pix_arr[index]))
				// Find minimum range and number of embeddable bits using range table
				min_range, bit_count := checkRangeTable(range_table, abs_pixel_difference)
				// Extract binary from difference
				secret := abs_pixel_difference - min_range
				secret_binary := ZeroLeftPad(strconv.FormatInt(int64(secret), 2), bit_count)
				extracted_binstring.WriteString(secret_binary)
				first_in_pair = true
			}
		}
	}
	// Convert to bytes
	output_bytes := BitstringToBytes(extracted_binstring.String())
	return output_bytes, nil
}

func RgbaToInt(s string) (int, error) {
	var colour int
	// Get R/G/B/A from first char
	first_char := s[0]
	switch first_char {
	case 'R':
		colour = Red
	case 'G':
		colour = Green
	case 'B':
		colour = Blue
	case 'A':
		colour = Alpha
	default:
		return -1, errors.New("invalid plane given. Only R/G/B/A are valid")
	}
	return colour, nil
}
