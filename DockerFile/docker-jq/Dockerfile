FROM ubuntu:14.04

RUN apt-get -y -qq update && \
	apt-get install -y -qq curl && \
	apt-get clean
#
# install jq to parse json within bash scripts
RUN curl -o /usr/local/bin/jq http://stedolan.github.io/jq/download/linux64/jq && \
  chmod +x /usr/local/bin/jq

# Define working directory.
WORKDIR /data

CMD ["-"]
ENTRYPOINT ["jq"]