FROM node:8.17.0-alpine

WORKDIR /home/microservice-app-example/frontend
COPY . /home/microservice-app-example/frontend

RUN npm install
RUN npm run build

EXPOSE 8080

CMD ["npm","start"]