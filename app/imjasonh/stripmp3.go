package imjasonh

import (
	"archive/zip"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// This script reads a zip file and writes a new zip file containing all the files in the input, without those ending with ".mp3"
// This is useful for getting zips from Google Takeout small enough to be uploaded to App Engine for processing.
func main() {
	flag.Parse()
	inFilename := flag.Arg(0)
	outFilename := flag.Arg(1)

	inZip, err := zip.OpenReader(inFilename)
	defer inZip.Close()
	if err != nil {
		log.Fatal(err)
	}

	outFile, err := os.Create(outFilename)
	defer outFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	outZip := zip.NewWriter(outFile)

	for _, inFile := range inZip.File {
		if strings.HasSuffix(inFile.Name, ".mp3") {
			log.Printf("Skipping %s", inFile.Name)
			continue
		}
		outFile, err := outZip.Create(inFile.Name)
		if err != nil {
			log.Fatal(err)
		}

		inRead, err := inFile.Open()
		defer inRead.Close()
		if err != nil {
			log.Fatal(err)
		}

		// TODO: Buffer instead of reading it all at once.
		data, err := ioutil.ReadAll(inRead)
		_, err = outFile.Write(data)
		if err != nil {
			log.Fatal(err)
		}
	}
}
