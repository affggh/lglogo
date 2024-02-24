package main

import (
	"flag"
	"lglogo"
	"log"
)

var (
	image     string
	picdir    string
	outimg    string
	isrepack  bool
	nopadding bool
)

func main() {
	flag.StringVar(&image, "image", "", "Logo image input")
	flag.StringVar(&picdir, "picdir", "pic", "Pic extract or repack dir")
	flag.StringVar(&outimg, "outimg", "new.img", "Repack output image name")
	flag.BoolVar(&isrepack, "repack", false, "Switch program mode to repack")
	flag.BoolVar(&nopadding, "nopadding", false, "Switch with no padding mode, if xoffset is zero but yoffset not")

	flag.Parse()

	if image == "" && !isrepack {
		log.Fatalln("Input image are not defined!")
	}

	if isrepack {
		lglogo.Repack(picdir, outimg, nopadding)
	} else {
		lglogo.Unpack(image, picdir)
	}
}
