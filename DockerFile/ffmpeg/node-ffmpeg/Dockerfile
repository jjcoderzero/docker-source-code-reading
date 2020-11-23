FROM shdr-applications-registry.cn-shanghai.cr.aliyuncs.com/base-image/node:10.15.0-buster-slim
RUN apt-get update \
  && apt-get install -y ffmpeg \

FROM shdr-applications-registry.cn-shanghai.cr.aliyuncs.com/application/shdr-push-live-streaming:node-ffmpeg
RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

COPY package.json /usr/src/app/
RUN npm install
COPY . /usr/src/app

CMD [ "npm", "start" ]

