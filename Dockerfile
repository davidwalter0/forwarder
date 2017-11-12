FROM debian:stretch-slim
COPY bin/forwarder /forwarder
CMD [ "/forwarder" ]
