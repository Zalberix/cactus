package core

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	dto "cactus/apps/core/internal/DTO"
	"cactus/apps/core/internal/error/validation"
	"cactus/apps/core/internal/http/request"
	"cactus/apps/core/internal/http/response"
)

type getFilesService interface {
	GetFile(ctx context.Context, uuid string) (dto.GetFile, error)
}

// Получение ссылки на страницу просмотра рассылки в реальном времени
func GetFile(s getFilesService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		queryValues := r.URL.Query()
		uuid := queryValues.Get("uuid")

		req := request.GetFileRequest{
			UUID: uuid,
		}

		// Переписать валидатор, так как сейчас мы пишем в структуре валидацию и приходится туда сюда прыгать.
		validationError, err := validation.ValidateStructure(&req)
		if err != nil {
			slog.Error("Ошибка структуры", slog.Any("err", err))
			response.FailJSON(w, "Ошибка при проверке полей, проверьте структуру.")
			return
		}
		if validationError != nil {
			response.ValidationJSON(w, "Ошибка валидации, проверьте отправляемые поля", validationError)
			return
		}

		file, err := s.GetFile(ctx, req.UUID)
		if err != nil {
			slog.Error(err.Error())
			response.FailJSON(w, "не возможно получить файл.")
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=file.txt")
		bfile, err := io.ReadAll(file.File)
		if err != nil {
			slog.Error(err.Error())
			response.FailJSON(w, "ошибка чтения файла.")
			return
		}

		w.Write(bfile)
	}
}
