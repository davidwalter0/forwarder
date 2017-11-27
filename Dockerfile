FROM davidwalter/debian-stretch-slim
VOLUME /var/lib/forwarder
COPY bin/forwarder /forwarder
COPY bin/echo /echo
EXPOSE 8888 80
CMD [ "/forwarder" ]
