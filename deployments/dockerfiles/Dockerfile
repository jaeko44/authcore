FROM debian:buster-slim
ARG COMMIT_REF_NAME=master
RUN apt-get update && apt-get install -y ca-certificates
WORKDIR /app
ADD authcore-dist-${COMMIT_REF_NAME}.tar.gz /app/
EXPOSE 7000 8000
ENTRYPOINT ["/app/entrypoint.sh"]