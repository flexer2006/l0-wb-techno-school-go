FROM confluentinc/cp-kafka:7.4.1

RUN mkdir -p /var/lib/kafka/data && \
    chown -R appuser:appuser /var/lib/kafka/data

CMD ["/etc/confluent/docker/run"]