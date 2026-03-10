package file

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

type Storage struct {
	FilesPath string
}

func NewFx() (*Storage, error) {
	return New("storages/local")
}

func New(basePath string) (*Storage, error) {
	binaryDir, err := os.Executable()
	if err != nil || strings.HasPrefix(filepath.Dir(binaryDir), os.TempDir()) {
		binaryDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("не удалось определить рабочую директорию: %w", err)
		}
	} else {
		binaryDir = filepath.Dir(binaryDir)
	}

	cleanBasePath := filepath.Clean(basePath)

	if filepath.IsAbs(cleanBasePath) || strings.HasPrefix(cleanBasePath, "..") {
		return nil, fmt.Errorf("недопустимый путь: %s", cleanBasePath)
	}

	fullPath := filepath.Join(binaryDir, cleanBasePath)

	relativePath, err := filepath.Rel(binaryDir, fullPath)
	if err != nil || strings.HasPrefix(relativePath, "../") {
		return nil, fmt.Errorf("путь выходит за пределы рабочей директории: %s", fullPath)
	}

	// Создание папки если ее нет, чтобы на момент сохранения файлов не создавать общую папку.
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		err = os.MkdirAll(fullPath, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("не удалось создать директорию: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("ошибка при проверке пути: %w", err)
	}

	return &Storage{
		FilesPath: fullPath,
	}, nil
}

// TODO доделать сохранение
func (s *Storage) Save(ctx context.Context, file multipart.File, ext mimetype.MIME) (path string, err error) {
	_ = ctx
	defer file.Close()

	bytes := make([]byte, 16) // TODO вынести в константы.
	_, err = rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("не смогли сгенерировать название файла: %w", err)
	}

	randomName := hex.EncodeToString(bytes)

	fileName := fmt.Sprintf("%s%s", randomName, ext.Extension())

	fullPath := filepath.Join(s.FilesPath, fileName)

	outFile, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("не смогли создать файла: %w", err)
	}
	defer outFile.Close()

	buffer := make([]byte, 2*1024)

	_, err = io.CopyBuffer(outFile, file, buffer)
	if err != nil {
		return "", fmt.Errorf("не смогли скопировать файла: %w", err)
	}
	return fileName, nil
}

func (s *Storage) Get(ctx context.Context, path string) (*os.File, error) {
	_ = ctx // TODO сделать открытие файла с контекстом

	fullPath := filepath.Join(s.FilesPath, path)

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл: %w", err)
	}

	return file, nil
}
