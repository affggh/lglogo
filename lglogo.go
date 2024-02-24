package lglogo

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"sync"
)

var Verbose = false

const LG_LOGO_IMAGE_HDR_OFFSET = 0

type LgLogoImageHdr struct {
	Magic    [16]uint8
	Num      uint32
	Unknow   uint32
	Metadata [16]uint8
	FSize    uint64
}

// Read data from a ReadSeeker
func (l *LgLogoImageHdr) Read(rd io.ReadSeeker) error {
	hsize := binary.Size(*l)
	rd.Seek(LG_LOGO_IMAGE_HDR_OFFSET, io.SeekStart)
	if hsize > 0 {
		return binary.Read(rd, binary.LittleEndian, l)
	}
	return nil
}

const LG_LOGO_IMAGE_METADATA_OFFSET = 0x1000
const LG_LOGO_IMAGE_MAX_METADATAS = 64

type LgLogoMetadata struct {
	Name    [40]uint8
	Offset  uint32
	Size    uint32
	Width   uint32
	Height  uint32
	XOffset uint32
	YOffset uint32
}

const LG_LOGO_IMAGE_DATA_OFFSET = 0x2000

func (l *LgLogoMetadata) Read(rd io.Reader) error {
	hsize := binary.Size(*l)
	if hsize > 0 {
		return binary.Read(rd, binary.LittleEndian, l)
	}
	return nil
}

// Strip string \x00
func Strip(input string) string {
	return strings.ReplaceAll(input, "\x00", "")
}

func Unpack(input string, outdir string) {
	fd, err := os.Open(input)
	if err != nil {
		log.Fatalln(err)
	}
	defer fd.Close()

	var (
		hdr      LgLogoImageHdr
		metadata LgLogoMetadata

		config Config

		wg sync.WaitGroup
	)

	info, err := os.Stat(outdir)
	if os.IsNotExist(err) {
		os.MkdirAll(outdir, 0777)
	} else {
		if !info.IsDir() {
			log.Fatalln("Cannot save to here:", info.Name())
		}
	}

	hdr.Read(fd)
	//fdhdr, err := os.Create(path.Join(outdir, "hdr"))
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//defer fdhdr.Close()

	config.Header.Magic = Strip(string(hdr.Magic[:]))
	config.Header.Metadata = Strip(string(hdr.Metadata[:]))
	config.Header.Unknow = int(hdr.Unknow)

	// Save hdr at pic
	//binary.Write(fdhdr, binary.LittleEndian, &hdr)

	fd.Seek(LG_LOGO_IMAGE_METADATA_OFFSET, io.SeekStart)

	parseImage := func(rd io.ReaderAt, outfile string, metadata LgLogoMetadata) {
		fdo, err := os.Create(outfile)
		if err != nil {
			log.Fatalln(err)
		}
		defer fdo.Close()

		rle := make([]byte, metadata.Size)
		rd.ReadAt(rle, int64(metadata.Offset))

		raw := Rle2Raw(rle)
		img := SwapRB(Raw2Image(raw, int(metadata.Width), int(metadata.Height)))

		png.Encode(fdo, img)

		defer wg.Done()
	}

	//fdmeta, err := os.Create("pic/metadata")
	for i := 0; i < int(hdr.Num); i++ {
		metadata.Read(fd)
		if Verbose {
			fmt.Printf("NUM\t\t: %d\n"+
				"\tName\t: %s\n"+
				"\tOffset\t: %d\n"+
				"\tSize\t: %d\n"+
				"\tWidth\t: %d\n"+
				"\tHeight\t: %d\n"+
				"\tXOffset\t: %d\n"+
				"\tYOffset\t: %d\n",
				i, metadata.Name, metadata.Offset, metadata.Size, metadata.Width, metadata.Height, metadata.XOffset, metadata.YOffset)
		} else {
			fmt.Printf("[%02d/%02d] Extract [%30s]...\n", i+1, hdr.Num, string(metadata.Name[:]))
		}
		wg.Add(1)
		go parseImage(fd, path.Join(outdir, strings.ReplaceAll(string(metadata.Name[:]), "\x00", "")+".png"), metadata)

		imagedata := ImageData{
			Name:    Strip(string(metadata.Name[:])),
			XOffset: int(metadata.XOffset),
			YOffset: int(metadata.YOffset),
		}

		config.ImageData = append(config.ImageData, imagedata)
	}

	wg.Wait()
	config.Saveto(path.Join(outdir, "info.toml"))

	fmt.Println("Done, image file save into dir:", outdir)
	fmt.Println("\tconfig file save info:", path.Join(outdir, "info.toml"))
}

// upper 0x1000
func UpperAlign(a uint32) uint32 {
	if a%0x1000 == 0 {
		return a // already align
	} else {
		return ((a >> 0xc) + 1) << 0xc
	}
}

func Repack(picdir string, outimg string, nopadding bool) {
	var (
		config Config

		hdr      LgLogoImageHdr
		metadata []LgLogoMetadata
	)

	fdconfig, err := os.Open(path.Join(picdir, "info.toml"))
	if err != nil {
		log.Fatalln(err)
	}
	defer fdconfig.Close()

	config.Read(fdconfig)

	copy(hdr.Magic[:], []uint8(config.Header.Magic))
	copy(hdr.Metadata[:], []uint8(config.Header.Metadata))

	hdr.Num = uint32(len(config.ImageData))

	fdo, err := os.Create(outimg)
	if err != nil {
		log.Fatalln(err)
	}
	defer fdo.Close()

	// Return rle data width and height
	parseImage := func(imagedata ImageData) ([]byte, int, int) {
		fdi, err := os.Open(path.Join(picdir, imagedata.Name+".png"))
		if err != nil {
			log.Fatalln(err)
		}
		defer fdi.Close()

		img, err := png.Decode(fdi)
		if err != nil {
			log.Fatalln(err)
		}
		img = SwapRB(img)

		width, height := img.Bounds().Dx(), img.Bounds().Dy()

		if !nopadding { // nopadding in some old devices
			// if XOffset is zero and background is black
			if imagedata.XOffset == 0 && imagedata.YOffset != 0 {
				// get first pixel of img
				r, g, b, _ := img.At(0, 0).RGBA()
				if r+g+b == 0 {
					newimg := image.NewRGBA(image.Rect(0, 0, width, height+imagedata.YOffset))
					draw.Draw(newimg, newimg.Bounds(), img, img.Bounds().Min, draw.Src)
					return Raw2Rle(Image2Raw(newimg), width, height+imagedata.YOffset), width, height
				}
			}
		}

		return Raw2Rle(Image2Raw(img), width, height), width, height
	}
	var imageData = config.ImageData

	offset := uint32(0x2000)
	for i := uint32(0); i < hdr.Num; i++ {
		fmt.Printf("[%02d/%02d] Write [%s]...\n", i+1, hdr.Num, imageData[i].Name)

		offset = UpperAlign(offset)
		rle, width, height := parseImage(imageData[i])
		m := LgLogoMetadata{}
		copy(m.Name[:], []uint8(imageData[i].Name))
		m.Width = uint32(width)
		m.Height = uint32(height)
		m.Offset = offset
		m.Size = uint32(len(rle))
		m.XOffset = uint32(imageData[i].XOffset)
		m.YOffset = uint32(imageData[i].YOffset)

		fdo.WriteAt(rle, int64(offset))

		metadata = append(metadata, m)
		offset += m.Size
	}

	hdr.Unknow = uint32(config.Header.Unknow)
	hdr.FSize = uint64(offset)
	// write hdr
	fdo.Seek(LG_LOGO_IMAGE_HDR_OFFSET, io.SeekStart)
	binary.Write(fdo, binary.LittleEndian, &hdr)

	// write metadata
	offset = LG_LOGO_IMAGE_METADATA_OFFSET
	fdo.Seek(int64(offset), io.SeekStart)
	for _, current := range metadata {
		binary.Write(fdo, binary.LittleEndian, &current)
		offset += uint32(binary.Size(metadata[0]))
	}

	fmt.Println("Done, new file save into:", outimg)

}
