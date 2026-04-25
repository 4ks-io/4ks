# tests

https://cloud.google.com/functions/docs/testing/test-event

# local dev

https://cloud.google.com/functions/docs/running/overview
https://cloud.google.com/functions/docs/running/function-frameworks
https://github.com/GoogleCloudPlatform//functions-framework-go#quickstart-hello-world-on-your-local-machine

# background functions

https://cloud.google.com/functions/docs/writing/write-event-driven-functions#background-functions
https://cloud.google.com/functions/docs/writing/write-event-driven-functions

# v2

gcloud auth application-default login

https://cloud.google.com/functions/docs/tutorials/storage

export FIRESTORE_PROJECT_ID=dev-4ks
export DISTRIBUTION_BUCKET="media-read.dev.4ks.io"
export FUNCTION_TARGET=UploadImage

go run cmd/main.go

```
curl localhost:8080 \
  -X POST \
  -H "Content-Type: application/json" \
  -H "ce-id: 123451234512345" \
  -H "ce-specversion: 1.0" \
  -H "ce-time: 2020-01-02T12:34:56.789Z" \
  -H "ce-type: google.cloud.storage.object.v1.finalized" \
  -H "ce-source: //storage.googleapis.com/projects/_/buckets/media-write.dev.4ks.io" \
  -H "ce-subject: objects/cfcpujq23akg00dj9u70.jpg" \
  -d '{
        "bucket": "media-write.dev.4ks.io",
        "contentType": "image/jpeg",
        "kind": "storage#object",
        "name": "cfcpujq23akg00dj9u70.jpg",
        "storageClass": "Standard",
        "timeCreated": "2020-04-23T07:38:57.230Z",
        "timeStorageClassUpdated": "2020-04-23T07:38:57.230Z",
        "updated": "2020-04-23T07:38:57.230Z"
      }'
```

gcloud config set project sbx-4ks
gcloud storage ls
gcloud storage ls --recursive gs://media.sbx.4ks.io/\*\*

## Runtime config

The media-upload function now loads config once during startup in `cmd/main.go`.

Required variables:

- `DISTRIBUTION_BUCKET`
- `FIRESTORE_PROJECT_ID`

Optional variables:

- `IO_4KS_DEVELOPMENT`
- `PORT`
