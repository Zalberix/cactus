package pkg

import (
	"fmt"
	"io"
	"mime/multipart"
	"slices"
	"strings"

	"github.com/gabriel-vasile/mimetype"

	"cactus/internal/http/request"
	"cactus/internal/service/core"
)

func ParseFileInMultipartForm(
	requestFiles []request.SendMessageRequestFile,
	formFiles map[string][]*multipart.FileHeader,
	allowedExtensions []string,
) ([]core.SetFileParams, string, error) {
	files := make([]core.SetFileParams, 0, len(requestFiles))
	for i, file := range requestFiles {
		key := fmt.Sprintf("Files.%v", i)
		MultipartFiles, ok := formFiles[file.Form]
		if !ok {
			return []core.SetFileParams{}, key, fmt.Errorf("файл по ключу %v не найден", file.Form)
		}
		ff, _ := MultipartFiles[0].Open()
		mtype, err := mimetype.DetectReader(ff)

		isAllowedExtensions := slices.Contains(allowedExtensions, mtype.Extension()[1:])

		if err != nil || !isAllowedExtensions {
			err := fmt.Errorf("файл может быть расширения %s", strings.Join(allowedExtensions, ", "))
			return []core.SetFileParams{}, key, err
		}

		_, err = ff.Seek(0, io.SeekStart)
		if err != nil {
			return []core.SetFileParams{}, key, fmt.Errorf("не удалось прочитать файл")
		}

		newFile := core.SetFileParams{
			File:  ff,
			Title: file.Title,
			Ext:   *mtype,
		}

		files = append(files, newFile)
	}

	return files, "", nil
}
