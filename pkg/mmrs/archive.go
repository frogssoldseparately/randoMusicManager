package mmrs

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/frogssoldseparately/simpleseek/sreader"
)

func MakeCreditedArchive(srcPath string, destPath string, bio *map[string]string) error {
	buffers, err := getExistingArchiveEntries(srcPath)
	if err != nil {
		return err
	}
	if _, ok := buffers["credits.txt"]; !ok {
		addCreditsToArchive(&buffers, bio, "title", "composers", "seq", "midi")
	}
	w, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer w.Close()
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()
	for name, buffer := range buffers {
		if err := writeNewArchiveEntry(zipWriter, name, buffer); err != nil {
			return err
		}
	}
	return nil
}

func writeNewArchiveEntry(zipWriter *zip.Writer, name string, buffer []byte) error {
	fileWriter, err := zipWriter.Create(name)
	if err != nil {
		return err
	}
	_, err = fileWriter.Write(buffer)
	return err
}

func addCreditsToArchive(buffers *map[string][]byte, bio *map[string]string, keys ...string) {
	var credits strings.Builder
	credits.WriteString("Credits:")
	for _, key := range keys {
		if val, ok := (*bio)[key]; ok && len(val) > 0 {
			fmt.Fprintf(&credits, "\n\t%s: %s", key, val)
		}
	}
	(*buffers)["credits.txt"] = []byte(credits.String())
}

func getExistingArchiveEntries(srcPath string) (map[string][]byte, error) {
	fileExt := filepath.Ext(srcPath)
	switch fileExt {
	case ".mmrs":
		return getArchiveComponents(srcPath)
	case ".zseq":
		fallthrough
	case ".aseq":
		fallthrough
	case ".seq":
		return getSequenceComponents(srcPath)
	default:
		return nil, fmt.Errorf("%s is not a supported extension", fileExt)
	}
}

func getSequenceComponents(srcPath string) (map[string][]byte, error) {
	fileExt := filepath.Ext(srcPath)
	fileBase := filepath.Base(srcPath)
	fileName := fileBase[0 : len(fileBase)-len(fileExt)]
	nameParts := strings.Split(fileName, "_")
	var categories string
	var bank string
	readPosition := len(nameParts) - 1
	if readPosition >= 2 {
		categories = nameParts[readPosition]
		readPosition--
	} else {
		categories = "bgm"
	}
	if readPosition >= 1 {
		bank = nameParts[readPosition]
	} else {
		bank = "3"
	}
	seqBuf, err := os.ReadFile(srcPath)
	if err != nil {
		return nil, err
	}
	buffers := map[string][]byte{
		fmt.Sprintf("%s.zseq", bank): seqBuf,
		"categories.txt":             []byte(categories),
	}
	return buffers, nil
}

func readArchiveEntryToBytes(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func getArchiveComponents(srcPath string) (map[string][]byte, error) {
	zipReader, err := sreader.OpenArchive(srcPath)
	if err != nil {
		return nil, err
	}
	defer zipReader.Close()
	buffers := map[string][]byte{}
	for _, file := range zipReader.GetFiles() {
		buf, err := readArchiveEntryToBytes(file)
		if err != nil {
			return nil, err
		}
		buffers[file.Name] = buf
	}
	return buffers, nil
}
