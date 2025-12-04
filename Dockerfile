## Builder ##
FROM golang:alpine AS builder

WORKDIR /build
COPY src/ .
RUN go mod tidy
RUN go build -ldflags="-s -w" -o ip-atlas.exe .

## Production image ##
FROM nginx:alpine
RUN apk update --no-cache && apk upgrade --no-cache
RUN apk add --no-cache bash curl cronie
RUN rm -rf /var/cache/apk/*

RUN mkdir /etc/cron
RUN echo "# empty line" > /etc/cron/crontab
RUN echo 'SHELL=/bin/sh' >> /etc/cron/crontab
RUN echo 'HOME=/app/ip-atlas' >> /etc/cron/crontab
RUN echo '* * * */3 * cd /app/ip-atlas/ && IP_ATLAS_PRODUCTION=TRUE ./ip-atlas.exe > /var/log/ip-atlas-cron.log 2>&1' >> /etc/cron/crontab
RUN echo "# empty line" >> /etc/cron/crontab
RUN crontab /etc/cron/crontab

RUN rm /etc/nginx/conf.d/default.conf
COPY nginx-config /etc/nginx/conf.d/ip-atlas.conf

COPY --from=builder /build/ip-atlas.exe /app/ip-atlas/ip-atlas.exe
COPY --from=builder /build/templates /app/ip-atlas/templates

WORKDIR /app/ip-atlas
ENV IP_ATLAS_PRODUCTION=TRUE

EXPOSE 23824

CMD ["/bin/sh", "-c", "/app/ip-atlas/ip-atlas.exe & crond -f & nginx -g 'daemon off;'"]