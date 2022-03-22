package lib

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"strconv"
)

const (
	Red int = iota
	Green
	Blue
	Alpha
)

func HasBit(n uint8, pos int) bool {
	// Checks if bit is set on int
	val := n & (1 << pos)
	return (val > 0)
}

func BitplaneArgsToArray(bitplane_args []string) ([][]interface{}, error) {
	/*
		Take input string, such as "R0 R1 B2" and convert to array ready for
		stego operations, i.e., "B0 B0 A2" -> [[Red, 0], [Blue, 0], [Alpha, 2]]
	*/
	var rgba_operations = make([][]interface{}, len(bitplane_args))
	for index, value := range bitplane_args {
		var colour int
		// Get R/G/B/A from first char
		first_char := value[0]
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
			return nil, fmt.Errorf("invalid colour input '%c' (must be R/G/B/A)", first_char)
		}
		// Get 0-7 from second char (can't be 7/9 as 8 bits per binary chunk)
		second_char := value[1]
		bitpos, err := strconv.Atoi(string(second_char))
		if err != nil {
			return nil, fmt.Errorf("invalid bitplane string '%s'. Should be in format 'R0', 'B7' etc", value)
		}
		if bitpos > 7 {
			return nil, fmt.Errorf("invalid bit position '%d'. Must be an int between 0-7", bitpos)
		}
		rgba_operations[index] = []interface{}{colour, bitpos}
	}
	return rgba_operations, nil
}

func FilepathToBitstream(secret_path string) ([]bool, error) {
	/*
		Open a file and convert to bool bitstream
	*/
	// Open secret file
	secret_file, err := os.Open(secret_path)
	if err != nil {
		return nil, err
	}
	defer secret_file.Close()

	// Get file details
	stats, statsErr := secret_file.Stat()
	if statsErr != nil {
		return nil, statsErr
	}
	var size int64 = stats.Size()

	// Create file read object
	bufr := bufio.NewScanner(secret_file)
	bufr.Split(bufio.ScanBytes)

	bitstream_arr := make([]bool, size*8)
	bit_pos := 0
	for bufr.Scan() {
		scanned_byte := bufr.Bytes()[0]
		for i := 0; i < 8; i++ {
			bitstream_arr[bit_pos*8+i] = int(scanned_byte>>uint(7-i)&0x01) == 1
		}
		bit_pos += 1
	}
	return bitstream_arr, nil
}

func BitstreamToBytes(bitstream []bool) []byte {
	/*
		Convert a bitstream string to bytes array.
	*/
	bytes_arr := make([]byte, (len(bitstream)+7)/8)
	for idx, val := range bitstream {
		if val {
			bytes_arr[idx/8] |= 0x80 >> uint(idx%8)
		}
	}
	return bytes_arr
}

func OpenImage(image_path string) (image.Image, error) {
	/*
		Open file as image
	*/
	// Open file
	imgfile, err := os.Open(image_path)
	if err != nil {
		return nil, err
	}
	defer imgfile.Close()

	// Parse as image
	imgfile.Seek(0, 0)
	img, _, err := image.Decode(imgfile)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func GetValuesPerPixel(img image.Image) (int, error) {
	/*
		Determine how many ints per pixel.
		i.e., RGB would have 3: R, G and B
			  RGBA would have 4: R, G, B and A
	*/
	var ints_per_pixel int
	switch img.ColorModel() {
	case color.GrayModel, color.Gray16Model, color.AlphaModel, color.Alpha16Model:
		ints_per_pixel = 1
	case color.RGBAModel, color.RGBA64Model, color.NRGBAModel, color.NRGBA64Model:
		ints_per_pixel = 4
	default:
		ints_per_pixel = -1
	}
	if ints_per_pixel == -1 {
		return 0, errors.New("invalid image type")
	}
	return ints_per_pixel, nil
}

func Abs(a int) int {
	if a < 0 {
		return a * -1
	}
	return a
}
