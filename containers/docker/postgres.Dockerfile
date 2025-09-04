FROM postgres:17.6

ENV PGDATA=/var/lib/postgresql/data/pgdata

CMD ["postgres"]