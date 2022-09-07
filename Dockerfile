FROM golang:1.19-alpine3.16 AS dev
#needs gdal
RUN apk add --no-cache \
    build-base \
    gcc \
    git

# TODO: add prod build needs gdal
# FROM osgeo/gdal:alpine-small-3.2.1 as prod
#WORKDIR /app
#COPY --from=dev /app/main .
#CMD [ "./main" ]