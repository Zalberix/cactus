package email

import (
	schemaApp "cactus/internal/schema"
	"context"
	"fmt"
	"net/http"

	"github.com/minio/minio-go/v7"
)

func (s *Service) SaveFiles(ctx context.Context, r *http.Request) ([]schemaApp.File, error) {

	file, handler, err := r.FormFile("go")
	if err != nil {
		return []schemaApp.File{}, fmt.Errorf("%v", "Ошибка получения файла из тела запроса")
	}
	defer file.Close()

	info, err := s.fileStorage.Storage.PutObject(ctx, "cactus", handler.Filename, file, handler.Size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return []schemaApp.File{}, fmt.Errorf("%v", err.Error())
	}
	//dst, err := os.Create(handler.Filename)
	//if err != nil {
	//	return []schemaApp.File{}, fmt.Errorf("%v", "Ошибка создания файла")
	//}
	//defer dst.Close()

	//if _, err := io.Copy(dst, file); err != nil {
	//	slog.Error(err.Error())
	//	w.Header().Set("Content-Type", "application/json")
	//	w.WriteHeader(403)
	//	jsonErrors, _ := json.Marshal(emailresponse.CreateResponseError("500.errors.upload_image.cannot_copy_to_file", nil))
	//	w.Write(jsonErrors)
	//	return
	//}

	return []schemaApp.File{
		schemaApp.File{
			Title:      "картинка",
			PathToFile: info.Key,
		},
	}, nil
}
