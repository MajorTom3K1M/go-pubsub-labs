FROM rabbitmq:3.13-management
RUN rabbitmq-plugins enable rabbitmq_stomp

EXPOSE 5672 15672 61613