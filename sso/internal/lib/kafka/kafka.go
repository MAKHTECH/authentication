package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"sso/sso/internal/domain/models"
)
import "github.com/IBM/sarama"

// Producer отправляет метрики в Kafka
type Producer struct {
	producer sarama.AsyncProducer
	//topic    string
}

func NewKafkaProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = false // Отключаем ожидание успешной отправки
	config.Producer.Return.Errors = true     // Логируем ошибки
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = sarama.CompressionSnappy // Оптимизация скорости
	config.Producer.Flush.Frequency = 500                  // Отправка сообщений пачками раз в 500мс

	fmt.Println(brokers)
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	go func() {
		for err = range producer.Errors() {
			log.Println("Ошибка отправки в Kafka:", err)
		}
	}()

	return &Producer{producer: producer}, nil
}

// SendMetric отправляет метрику в Kafka
func (kp *Producer) SendMetric(metric map[string]interface{}, topic models.Topic) {
	message, err := json.Marshal(metric)
	if err != nil {
		log.Println("Ошибка сериализации метрики:", err)
		return
	}

	// Проверяем, что продюсер жив и канал открыт
	if kp.producer == nil {
		log.Println("Продюсер не инициализирован")
		return
	}

	select {
	case kp.producer.Input() <- &sarama.ProducerMessage{
		Topic: string(topic),
		Value: sarama.ByteEncoder(message),
	}:
		log.Println("Метрика отправлена в Kafka:", string(message))

	case err, ok := <-kp.producer.Errors():
		if !ok {
			log.Println("Канал ошибок закрыт")
		} else if err != nil {
			log.Println("Ошибка отправки в Kafka:", err)
		} else {
			log.Println("Получено nil из канала ошибок")
		}

	default:
		log.Println("Не удалось отправить метрику: канал блокируется или закрыт")
	}
}

// Close закрывает продюсер
func (kp *Producer) Close() {
	if err := kp.producer.Close(); err != nil {
		fmt.Println("Ошибка при закрытии Kafka продюсера:", err)
	}
}
