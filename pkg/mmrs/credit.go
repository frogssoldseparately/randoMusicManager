package mmrs

import (
	"archive/zip"
	"fmt"
	"io"
	"os"

	"github.com/frogssoldseparately/simpleseek/sreader"
)

func AddCredits(mmrsPath string, bio *map[string]string) error {
	buffers, err := getExistingFiles(mmrsPath)
	if err != nil {
		return err
	}
	if _, ok := buffers["credits.txt"]; ok {
		return fmt.Errorf("this archive already has a credits file")
	}
	w, err := os.Create(mmrsPath)
	if err != nil {
		return err
	}
	defer w.Close()
	zipW := zip.NewWriter(w)
	defer zipW.Close()
	for fileName, buf := range buffers {
		if err := writeFile(zipW, fileName, buf); err != nil {
			return err
		}
	}
	zipCreditW, err := zipW.Create("credits.txt")
	if err != nil {
		return err
	}
	writeBioEntry(zipCreditW, bio, "title", "composers", "seq", "midi")
	return nil
}

func writeFile(zipW *zip.Writer, fileName string, buffer []byte) error {
	fileWriter, err := zipW.Create(fileName)
	if err != nil {
		return err
	}
	_, err = fileWriter.Write(buffer)
	return err
}

func getExistingFiles(path string) (map[string][]byte, error) {
	zipR, err := sreader.OpenArchive(path)
	if err != nil {
		return nil, err
	}
	defer zipR.Close()
	buffers := map[string][]byte{}
	for _, file := range zipR.GetFiles() {
		if buf, err := getFileBytes(file); err != nil {
			return nil, err
		} else {
			buffers[file.Name] = buf
		}
	}
	return buffers, nil
}

func getFileBytes(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func writeBioEntry(w io.Writer, bio *map[string]string, keys ...string) {
	for _, key := range keys {
		if val, ok := (*bio)[key]; ok && len(val) > 0 {
			fmt.Fprintf(w, "%s: %s\n", key, val)
		}
	}
}
