package mmrs

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ConvertToMmrs(srcFilePath string, destPath string) error {
	fileExt := filepath.Ext(srcFilePath)
	fileBasename := filepath.Base(srcFilePath)
	sequenceInfo := fileBasename[0 : len(fileBasename)-len(fileExt)]
	nameParts := strings.Split(sequenceInfo, "_")
	catString := nameParts[len(nameParts)-1]
	bankString := nameParts[len(nameParts)-2]

	w, err := os.Create(destPath)
	if err != nil {
		return err
	}
	zipW := zip.NewWriter(w)
	zipSeqW, err := zipW.Create(fmt.Sprintf("%s.zseq", bankString))
	if err != nil {
		return err
	}
	seqBuf, err := os.ReadFile(srcFilePath)
	if err != nil {
		return err
	}
	zipSeqW.Write(seqBuf)
	zipCatW, err := zipW.Create("categories.txt")
	fmt.Fprintf(zipCatW, "%s", catString)
	zipW.Close()
	w.Close()
	return nil
}
