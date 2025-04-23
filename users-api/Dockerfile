FROM openjdk:8-jdk

ENV JWT_SECRET=PRFT
ENV SERVER_PORT=8083

COPY . .
RUN sed -i 's/\r$//' mvnw
RUN ./mvnw clean install

ENTRYPOINT java -jar target/users-api-0.0.1-SNAPSHOT.jar