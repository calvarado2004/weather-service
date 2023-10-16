package main

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQService interface {
	Dial(url string) (AMQPConnection, error)
	Channel(conn AMQPConnection) (Channel, error)
	QueueDeclare(ch Channel, name string) (amqp.Queue, error)
	Publish(ch Channel, queueName string, body []byte) error
}

type AMQPConnection interface {
	Close() error
	Channel() (Channel, error)
}

type RealRabbitMQService struct{}

type Channel interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Close() error
}

type RabbitConnectionWrapper struct {
	conn *amqp.Connection
}

func (rcw *RabbitConnectionWrapper) Close() error {
	return rcw.conn.Close()
}

func (rcw *RabbitConnectionWrapper) Channel() (Channel, error) {
	ch, err := rcw.conn.Channel()
	if err != nil {
		return nil, err
	}
	return &RabbitChannelWrapper{ch: ch}, nil
}

type RabbitChannelWrapper struct {
	ch *amqp.Channel
}

func (rcw *RabbitChannelWrapper) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	return rcw.ch.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)
}

func (rcw *RabbitChannelWrapper) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return rcw.ch.PublishWithContext(ctx, exchange, key, mandatory, immediate, msg)
}

func (rcw *RabbitChannelWrapper) Close() error {
	return rcw.ch.Close()
}

func (r *RealRabbitMQService) Dial(url string) (AMQPConnection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	return &RabbitConnectionWrapper{conn: conn}, nil
}

func (r *RealRabbitMQService) Channel(conn AMQPConnection) (Channel, error) {
	return conn.Channel()
}

func (r *RealRabbitMQService) QueueDeclare(ch Channel, name string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		name,  // Name of the queue
		false, // Durable
		false, // Delete when unused
		false, // Exclusive
		false, // No-wait
		nil,   // Arguments
	)
}

func (r *RealRabbitMQService) Publish(ch Channel, queueName string, body []byte) error {
	return ch.PublishWithContext(
		context.TODO(),
		"",        // Exchange
		queueName, // Routing key
		false,     // Mandatory
		false,     // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
