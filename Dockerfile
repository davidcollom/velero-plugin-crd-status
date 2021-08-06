FROM debian:stretch-slim
RUN mkdir /plugins
ADD velero-* /plugins/
USER nobody:nogroup
ENTRYPOINT ["/bin/bash", "-c", "cp /plugins/* /target/."]
