# image-resize-proxy
A http server that returns images from google storage applying resizing and compress.


## How to run
```
go build
export GOOGLE_APPLICATION_CREDENTIALS="{google_credentials_path.json}"
./image-resize-proxy
```

## How to request
```
curl --location --request POST 'http://localhost:8080/google/bucket/download' \
--header 'Content-Type: application/json' \
--data-raw '{
    "bucketName": "{bucketName}",
    "filename": "{filepath}",
    "maxWidth": {maxWidthAllowed}
}'

```
