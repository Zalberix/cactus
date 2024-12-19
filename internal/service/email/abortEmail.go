package email

import (
	"context"
	"fmt"
)

func (s *Service) AbortEmail(ctx context.Context, uuid string) (string, error) {
	// TODO: после добавления redis дополнить метод удалением из очереди в redis, а также убрать из очереди на отслеживание через сокет.
	// TODO: Разобраться откуда берутся шаги для pipline так как сейчас их можно взять только из кода. А также надо разобраться какие шагие есть в системе и какие можно откатить.
	// str, err := s.storage.AbortMessageByUUID(ctx, uuid)
	return "", fmt.Errorf("невозможно отменить сообщение")
}
