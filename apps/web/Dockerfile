# # stage: deps (base)
# FROM node:25.9.0-alpine AS deps
# RUN addgroup -S appgroup && adduser -S appuser -G appgroup
# RUN npm install -g pnpm@10.28.0

# WORKDIR /code
# COPY /PACKAGE_JSON /code
# COPY /pnpm* .
# COPY /.pnpmfile.cjs .
# RUN pnpm install --frozen-lockfile && rm -rf /code
#   # output -> /root/.local/share/pnpm/store/v3

# stage: build
FROM node:25.9.0-alpine@sha256:bdf2cca6fe3dabd014ea60163eca3f0f7015fbd5c7ee1b0e9ccb4ced6eb02ef4 AS build
LABEL org.opencontainers.image.source="https://github.com/4ks-io/4ks"
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN npm install -g pnpm@10.28.0

WORKDIR /code
COPY /PACKAGE_JSON /code
COPY /pnpm* .
COPY /.pnpmfile.cjs .
RUN pnpm install --frozen-lockfile

COPY ./libs/reserved-words ./apps/web/public
COPY ./libs/ts/api-fetch /code/libs/ts/api-fetch
RUN pnpm --filter @4ks/api-fetch build

WORKDIR /code/apps/web
COPY ./apps/web /code/apps/web
RUN npx next build
  # output -> /app/apps/web/.next

# EXPOSE 3000
# WORKDIR /code/apps/web
# CMD [ "npm", "run", "start" ]

# runtime
FROM node:25.9.0-alpine@sha256:bdf2cca6fe3dabd014ea60163eca3f0f7015fbd5c7ee1b0e9ccb4ced6eb02ef4
LABEL org.opencontainers.image.source="https://github.com/4ks-io/4ks"
RUN mkdir /home/appuser && chmod 777 /home/appuser
USER appuser

COPY --from=build /etc/passwd                      /etc/passwd
COPY --from=build /root/.local/share/pnpm          /root/.local/share/pnpm
COPY --from=build /code/package.json               /app/package.json
COPY --from=build /code/node_modules               /app/node_modules
COPY --from=build /code/apps/web/package.json /app/apps/web/package.json
COPY --from=build /code/apps/web/node_modules /app/apps/web/node_modules
COPY --from=build /code/apps/web/public       /app/apps/web/public
COPY --from=build /code/apps/web/.next        /app/apps/web/.next

EXPOSE 3000
WORKDIR /app/apps/web
CMD [ "npx", "next", "start" ]
