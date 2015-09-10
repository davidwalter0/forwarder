FROM centos:latest
COPY simple-forwarder /simple-forwarder
CMD [ "/simple-forwarder" ]
