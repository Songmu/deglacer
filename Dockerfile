FROM golang:1.17.3-alpine AS build

RUN apk add --update --no-cache make git ca-certificates openssl

ENV APP_ROOT /deglacer/
WORKDIR $APP_ROOT
COPY . $APP_ROOT

RUN make build

FROM golang:alpine

COPY --from=build /deglacer/bin /bin

ENV PORT 8080
ENV SLACK_TOKEN ""
ENV SLACK_SIGNING_SECRET ""
ENV KIBELA_TOKEN ""
ENV KIBELA_TEAM ""

EXPOSE $PORT
ENTRYPOINT /bin/deglacer
