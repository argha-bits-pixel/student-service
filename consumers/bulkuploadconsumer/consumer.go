package bulkuploadconsumer

import (
	"encoding/json"
	"log"
	"student-service/adapters/rabbitmq"
	"student-service/consumers"
	"student-service/controller/bulkuploadcontroller"
	"student-service/models"
	"student-service/usecase"

	"github.com/streadway/amqp"
)

type BulkUploadConsumer struct {
	BulkuploadUc usecase.BulkUploadRequestHandler
	Rabbit       *rabbitmq.RabbitChannel
}

func (b *BulkUploadConsumer) SubscribeToBulkUploadQueue(queueName string) {
	forever := make(chan bool)
	channel, err := b.Rabbit.Consume(queueName, "", false, false, false, false, amqp.Table{})
	if err != nil {
		log.Fatalf("Unable to start Processing from queue %s", err.Error())
	}

	for j := range channel {
		//UnMarshall and see what kind of data
		data := new(models.BulkUploadModel)
		if err = json.Unmarshal(j.Body, data); err != nil {
			log.Println("Unable to Unmarshall Data Packet for processing ", err.Error())
			j.Ack(false)
		}
		switch data.RequestType {
		case bulkuploadcontroller.BULK_STUDENT_RECORD:
			go b.BulkuploadUc.ProcessBulkStudentRecord(data)
			j.Ack(false)
		case bulkuploadcontroller.BULK_VACCINE_RECORD:
			go b.BulkuploadUc.ProcessBulkVaccineRecord(data)
			j.Ack(false)
		default:
			log.Println("Unknown Request Type")
		}
	}
	<-forever
}

func NewBulkUploadProcessor(bulkuploadUc usecase.BulkUploadRequestHandler, rabb *rabbitmq.RabbitChannel) consumers.BulkUploadConsumers {
	return &BulkUploadConsumer{BulkuploadUc: bulkuploadUc, Rabbit: rabb}
}
