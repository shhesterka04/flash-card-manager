package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"flash-card-manager/internal/infrastructure/kafka"
	"flash-card-manager/pkg/logger"
	"io/ioutil"
	"net/http"

	"github.com/IBM/sarama"
)

func ReadAndRestoreRequestBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, errors.New("request body is nil")
	}

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	return bodyBytes, nil
}

type GRPCKafkaEventMatcher struct {
	ExpectedType  string
	ExpectedQuery string
}

func (k *GRPCKafkaEventMatcher) Matches(x interface{}) bool {
	msg, ok := x.(*sarama.ProducerMessage)
	if !ok {
		logger.GetLogger().Info("GRPCKafkaEventMatcher: Тип сообщения не соответствует ProducerMessage")
		return false
	}

	valueBytes, err := msg.Value.Encode()
	if err != nil {
		logger.GetLogger().Sugar().Infof("GRPCKafkaEventMatcher: Ошибка кодирования значения сообщения: %v", err)
		return false
	}

	var actualEvent kafka.Event
	if err := json.Unmarshal(valueBytes, &actualEvent); err != nil {
		logger.GetLogger().Sugar().Infof("GRPCKafkaEventMatcher: Ошибка кодирования значения сообщения: %v", err)
		return false
	}

	return actualEvent.Type == k.ExpectedType && actualEvent.Query == k.ExpectedQuery
}

func (k *GRPCKafkaEventMatcher) String() string {
	return fmt.Sprintf("ожидается событие типа '%s' с запросом '%s'", k.ExpectedType, k.ExpectedQuery)
}
