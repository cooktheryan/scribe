FROM centos:8

ARG restic_version=0.11.0

RUN yum update -y && \
    yum install -y \
      bash \
      bzip2 \
    && yum clean all && \
    rm -rf /var/cache/yum

COPY active.sh \
     /

RUN chmod a+rx /active.sh

RUN curl -SL \
    https://github.com/restic/restic/releases/download/v"${restic_version}"/restic_"${restic_version}"_linux_amd64.bz2 \
    -o /tmp/restic_${restic_version}_linux_amd64.bz2 && \
    bzip2 -dc /tmp/restic_${restic_version}_linux_amd64.bz2 > /usr/local/sbin/restic && \
    chmod +x /usr/local/sbin/restic

ARG builddate_arg="(unknown)"
ARG version_arg="(unknown)"
ENV builddate="${builddate_arg}"
ENV version="${version_arg}"

LABEL org.label-schema.build-date="${builddate}" \
      org.label-schema.description="restic-based data mover for Scribe" \
      org.label-schema.license="AGPL v3" \
      org.label-schema.name="scribe-mover-restic" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.vcs-ref="${version}" \
      org.label-schema.vcs-url="https://github.com/backube/scribe" \
      org.label-schema.vendor="Backube" \
      org.label-schema.version="${version}"

ENTRYPOINT [ "/bin/bash" ]
