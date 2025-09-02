FROM confluentinc/cp-zookeeper:7.4.1

HEALTHCHECK --interval=10s --timeout=5s --start-period=60s --retries=5 \
  CMD ["bash", "-lc", "echo ruok | nc -w 3 127.0.0.1 2181 >/dev/null 2>&1 || exit 1"]

CMD ["bash", "-c", "/etc/confluent/docker/run"]