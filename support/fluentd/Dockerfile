FROM fluent/fluentd:v0.12.29-onbuild

WORKDIR /home/fluent
ENV PATH /home/fluent/.gem/ruby/2.3.0/bin:$PATH

USER root
RUN apk --no-cache --update add sudo build-base ruby-dev && \

    sudo -u fluent gem install fluent-plugin-secure-forward && \
    sudo -u fluent gem install fluent-plugin-influxdb && \
    sudo -u fluent gem install fluent-plugin-concat && \
    sudo -u fluent gem install fluent-plugin-s3 && \
    sudo -u fluent gem install fluent-plugin-forest && \
    sudo -u fluent gem install fluent-plugin-cloudwatch-logs && \

    rm -rf /home/fluent/.gem/ruby/2.3.0/cache/*.gem && sudo -u fluent gem sources -c && \
    apk del sudo build-base ruby-dev && rm -rf /var/cache/apk/*

EXPOSE 24284

USER fluent
CMD exec fluentd -c /fluentd/etc/$FLUENTD_CONF -p /fluentd/plugins $FLUENTD_OPT