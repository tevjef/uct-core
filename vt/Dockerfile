FROM openjdk:8-jdk-slim AS build-env

COPY . /build

RUN cd /build && ./gradlew fatJar

FROM openjdk:8-jre-alpine
RUN apk --no-cache add ca-certificates openssl && \
    update-ca-certificates
COPY --from=build-env /build/server/build/libs /
EXPOSE 13000
CMD ["java", "-jar", "server-all.jar"]
