package main

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"lglogo"
	"log"
	"os"
)

func compare(a []uint8, b []uint8) bool {
	if len(a) == len(b) {
		for index, current := range a {
			if current != b[index] {
				log.Printf("Not same at %08x\n", current)
				return false
			}
		}
		return true
	} else {
		log.Println("Different size!")
		return false
	}
}

func rlelength(input []byte) int {
	rd := bytes.NewReader(input)
	var (
		output_size = 0
		block       = lglogo.RunLengthBlock{}
	)
	// calc size
	for i := 0; i < len(input); i += 4 {
		block.Read(rd)
		output_size += (int(block.Length) * 3) // RGB
	}

	return output_size
}

func verify() {
	fd, err := os.Open("raw_resources_a.img")
	if err != nil {
		log.Fatalln(err)
	}

	var (
		hdr      lglogo.LgLogoImageHdr
		metadata lglogo.LgLogoMetadata
	)

	hdr.Read(fd)
	fd.Seek(lglogo.LG_LOGO_IMAGE_METADATA_OFFSET, io.SeekStart)

	for i := uint32(0); i < hdr.Num; i++ {
		metadata.Read(fd)

		fmt.Printf("Name: %s\n", metadata.Name)

		data := make([]byte, metadata.Size)
		fd.ReadAt(data, int64(metadata.Offset))

		rawlen := len(lglogo.Rle2Raw(data))
		fmt.Println("data size  :", rawlen)
		fmt.Println("data calc  :", metadata.Width*metadata.Height*3)
		if rawlen != int(metadata.Width*metadata.Height*3) {
			fmt.Printf("%#v\n", metadata)
		}
	}
}

func main() {
	fd, err := os.Open("pic/load_charger_image.png")
	if err != nil {
		log.Fatalln(err)
	}
	defer fd.Close()

	img, err := png.Decode(fd)
	if err != nil {
		log.Fatalln(err)
	}

	raw := lglogo.Image2Raw(img)
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	rle := lglogo.Raw2Rle(raw, width, height)

	fdt, err := os.Open("t.rle")
	if err != nil {
		log.Fatalln(err)
	}
	defer fdt.Close()

	fdo, err := os.Create("t2.rle")
	if err != nil {
		log.Fatalln(err)
	}
	defer fdo.Close()

	fdo.Write(rle)
	test, _ := io.ReadAll(fdt)
	log.Println("origin  size: ", rlelength(test))
	log.Println("new     size: ", rlelength(rle))
	log.Println("correct size: ", width*height*3)
	if compare(rle, test) {
		log.Println("Success")
	} else {
		log.Println("Failed")
	}

	verify()
}
