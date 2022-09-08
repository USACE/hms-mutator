FROM osgeo/gdal:alpine-normal-3.2.1 AS dev

RUN apk add --no-cache \
    build-base \
    gcc \
    git

# TODO: add prod build needs gdal
# FROM osgeo/gdal:alpine-small-3.2.1 as prod
#WORKDIR /app
#COPY --from=dev /app/main .
#CMD [ "./main" ]