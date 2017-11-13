FROM debian:stretch-slim
COPY bin/forwarder /forwarder
RUN apt update; apt install -yqq dnsutils; apt-get clean;
VOLUME /var/lib/forwarder
CMD [ "/forwarder" ]
