# Rabbit-consumer

This is a monorepo which subscribes to a RabbitMQ topic exchange and can then process the events in an arbitrary way. To add a new kind of consumer you just need to implement the following interface:

```
func(d rabbitmq.Delivery) rabbitmq.Action 
```

and define a new `..Handler` constant.

# Configuration

- `RABBITMQ_URI`: e.g. `amqp://user:password@host/vhost`
- `RABBITMQ_EXCHANGE_NAME`: which exchange to subscribe to
- `RABBITMQ_ROUTING_KEY`: routing key to subscribe to
