package lglogo

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"io"
)

type RGB struct {
	R uint8
	G uint8
	B uint8
}

type RunLengthBlock struct {
	Length uint8
	Pixel  RGB
}

func (r *RunLengthBlock) Read(rd io.Reader) error {
	hsize := binary.Size(*r)
	if hsize > 0 {
		return binary.Read(rd, binary.LittleEndian, r)
	}
	return nil
}

func Rle2Raw(input []uint8) []uint8 {
	rd := bytes.NewReader(input)
	var (
		output_size = 0
		block       = RunLengthBlock{}
	)
	// calc size
	for i := 0; i < len(input); i += 4 {
		block.Read(rd)
		output_size += (int(block.Length) * 3) // RGB
	}

	output := make([]byte, output_size)
	offset := 0

	rd.Seek(0, io.SeekStart)
	for i := 0; i < len(input); i += 4 {
		block.Read(rd)
		for j := 0; j < int(block.Length); j++ {
			output[offset] = block.Pixel.R
			output[offset+1] = block.Pixel.G
			output[offset+2] = block.Pixel.B
			offset += 3
		}
	}

	return output
}

func Raw2Image(input []byte, width int, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

	var index int
	var r, g, b byte
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// 计算 rawImage 中对应像素的索引
			index = (y*width + x) * 3

			// 从 rawImage 中获取 RGB 值
			r = input[index]
			g = input[index+1]
			b = input[index+2]

			// 设置图像的像素
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	return img
}

// Convert image.Image into a 24bit RGB raw image
func Image2Raw(input image.Image) []byte {
	bounds := input.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	raw := make([]byte, width*height*3)

	var index int
	var r, g, b uint32
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			index = (y*width + x) * 3

			// drop alpha channal
			r, g, b, _ = input.At(x, y).RGBA()

			raw[index] = uint8(r >> 8)
			raw[index+1] = uint8(g >> 8)
			raw[index+2] = uint8(b >> 8)

		}
	}

	return raw
}

// run_length r g b
func Raw2Rle(input []byte, width int, height int) []byte {
	size := width * height

	var pixel, last RGB
	var count uint8 = 0

	buf := new(bytes.Buffer)
	// calc input
	for i := 0; i < size; i++ {
		pixel = RGB{input[i*3], input[i*3+1], input[i*3+2]}
		if i == 0 {
			last = pixel
			count++
		} else {
			if pixel == last {
				if count == 255 {
					buf.WriteByte(count)
					binary.Write(buf, binary.BigEndian, &last)
					count = 1
				} else {
					count++
				}
			} else {
				buf.WriteByte(count)
				binary.Write(buf, binary.BigEndian, &last)
				count = 1
			}
			if i == size-1 {
				last = pixel
				buf.WriteByte(count)
				binary.Write(buf, binary.BigEndian, &last)
			}
		}
		last = pixel
	}
	return buf.Bytes()
}

func SwapRB(input image.Image) image.Image {
	bounds := input.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			old := input.At(x, y)
			r, g, b, a := old.RGBA()

			new := color.RGBA{
				R: uint8(b >> 8),
				G: uint8(g >> 8),
				B: uint8(r >> 8),
				A: uint8(a >> 8),
			}

			img.SetRGBA(x, y, new)
		}
	}

	return img
}
