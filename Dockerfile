FROM davidwalter/debian-stretch-slim
VOLUME /var/lib/forwarder
COPY bin/forwarder /forwarder
CMD [ "/forwarder" ]
