FROM confluentinc/cp-kafka:7.4.1

RUN mkdir -p /var/lib/kafka/data && \
    chown -R appuser:appuser /var/lib/kafka/data

HEALTHCHECK --interval=10s --timeout=5s --start-period=60s --retries=5 \
  CMD ["bash", "-lc", "/usr/bin/kafka-broker-api-versions --bootstrap-server 127.0.0.1:9092 >/dev/null 2>&1 || exit 1"]

CMD ["/etc/confluent/docker/run"]