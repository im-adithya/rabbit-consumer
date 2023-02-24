# Rabbit-consumer

This is a monorepo which subscribes to a RabbitMQ topic exchange and can then process the events in an arbitrary way. To add a new kind of consumer you just need to implement the Handler interface:
```
type Handler interface {
	Handle(ctx context.Context, msg amqp.Delivery) error
}
```
and define a new `..Handler` constant.

# Configuration

- `HANDLER_TYPE`: currently only `print_handler` and `slack_raw_invoice_handler` are defined
- `RABBITMQ_URI`: e.g. `amqp://user:password@host/vhost`
- `RABBITMQ_EXCHANGE_NAME`: which exchange to subscribe to
- `RABBITMQ_QUEUE_NAME`: queue name, if not defined it is defined as `<EXCHANGE_NAME>_<HANDLER_TYPE>`
- `RABBITMQ_ROUTING_KEY`: routing key to subscribe to

# Slack invoice publisher
This handler listens to LND invoices and currently examines if someone send us a keysend payment without specifying the recipient. In that case it will post the boostagram information to a Slack channel so we can find the podcaster or app that might have something misconfigured.

- `SLACK_WEBHOOK_URL`
- `SLACK_CHANNEL`
# Print handler
Can be used for debugging just logs the raw events, does nothing else
