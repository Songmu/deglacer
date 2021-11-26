FROM golang:1.17 AS build
RUN apt-get update
RUN apt-get install -y make git ca-certificates openssl file
ENV APP_ROOT /deglacer/
WORKDIR $APP_ROOT
COPY . $APP_ROOT
ARG BUILD_VERSION=unknown
ARG BUILD_HASH=unknown
RUN make build

FROM gcr.io/distroless/base-debian11
COPY --from=build /deglacer/bin/ /bin
ENV PORT 8080
ENV SLACK_TOKEN ""
ENV SLACK_SIGNING_SECRET ""
ENV KIBELA_TOKEN ""
ENV KIBELA_TEAM ""
EXPOSE $PORT
ENTRYPOINT ["/bin/deglacer"]
