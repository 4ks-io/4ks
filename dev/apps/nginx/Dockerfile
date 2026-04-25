FROM nginx:mainline-alpine@sha256:5616878291a2eed594aee8db4dade5878cf7edcb475e59193904b198d9b830de
LABEL org.opencontainers.image.source="https://github.com/4ks-io/4ks"

RUN chmod -R a+w /var/cache/nginx/ \
        && cat > /var/run/nginx.pid \
        && chmod a+w /var/run/nginx.pid \
        && rm /etc/nginx/conf.d/*

COPY info.html /usr/share/nginx/html/
COPY 502.html /usr/share/nginx/html/

EXPOSE 8080
EXPOSE 4433
USER nginx
