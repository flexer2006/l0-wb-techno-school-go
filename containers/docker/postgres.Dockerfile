FROM postgres:17.6

ENV PGDATA=/var/lib/postgresql/data/pgdata

HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=5 \
  CMD ["sh", "-c", "pg_isready -U \"$POSTGRES_USER\" -d \"$POSTGRES_DB\" || exit 1"]

CMD ["postgres"]