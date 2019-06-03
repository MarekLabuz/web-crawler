FROM node:11

EXPOSE 80
EXPOSE 8888

RUN mkdir -p /usr/src/app
RUN mkdir -p /run/attachments
WORKDIR /usr/src/app

COPY package.json package.json

RUN npm install && npm cache clean --force

COPY . .

CMD ["npm", "start"]
