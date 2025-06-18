# Configure Storage

## Local Storage
By default, homebox uses local storage at the `./data` relative path to the binary, or `/data` in the docker container.
You can change the storage path by setting the `HBOX_STORAGE_CONN_STRING` to `file://path/you/want`. The `HBOX_STORAGE_PREFIX_PATH`
can be used to set a "prefix" for the storage. This "prefix" comes after the path in the connection string.

::: warning
  The local storage path must be writable by the user running the homebox process. Homebox will automatically create the directory for `file://./` but if you specify a different path, you must ensure that the directory exists and is writable.
:::
## S3 Storage

### Authentication
To authenticate with S3, you will need to set the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables. Optionally, you can also set `AWS_SESSION_TOKEN` if you are using temporary credentials.

### AWS S3
You can use S3 storage by setting the `HBOX_STORAGE_CONN_STRING` to `s3://my-bucket?region=region-name&awssdk=v2`.

In this case, the `HBOX_STORAGE_PREFIX_PATH` can be used to set a "prefix" for the storage. This "prefix" comes after the bucket name in the connection string.

### S3-Compatible Storage
You can also use S3-compatible storage by setting the `HBOX_STORAGE_CONN_STRING` to `s3://my-bucket?awssdk=v2&endpoint=http://my-s3-compatible-endpoint.tld&disableSSL=true&s3ForcePathStyle=true`.

This allows you to connect to S3-compatible services like MinIO, DigitalOcean Spaces, or any other service that supports the S3 API. Configure the `disableSSL`, `s3ForcePathStyle`, and `endpoint` parameters as needed for your specific service.

### Extra Connection Parameters
Additionally, the parameters in the URL can be used to configure specific S3 settings:
- `region`: The AWS region where the bucket is located.
- `awssdk`: The version of the AWS SDK to use (e.g., `v2`). (We highly recommend using `v2` for better performance and features.)
- `endpoint`: The custom endpoint for S3-compatible storage services.
- `s3ForcePathStyle`: Whether to force path-style access (set to `true` or `false`).
- `disableSSL`: Whether to disable SSL (set to `true` or `false`).
- `sseType`: The server-side encryption type (e.g., `AES256` or `aws:kms` or `aws:kms:dsse`).
- `kmskeyid`: The KMS key ID for server-side encryption.
- `fips`: Whether to use FIPS endpoints (set to `true` or `false`).
- `dualstack`: Whether to use dual-stack endpoints (set to `true` or `false`).
- `accelerate`: Whether to use S3 Transfer Acceleration (set to `true` or `false`).


## Google Cloud Storaget

### Authentication
To authenticate with Google Cloud Storage, you will need to set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to the path of your service account key file.
This file should be in JSON format and contain the necessary credentials to access your Google Cloud Storage bucket and must be made available to the application if running docker via volume read-only mounts.

### Using Google Cloud Storage
You can use Google Cloud Storage by setting the `HBOX_STORAGE_CONN_STRING` to `gcs://my-bucket`.

## Azure Blob Storage
### Authentication
To authenticate with Azure blob storage, you will need to set the `AZURE_STORAGE_ACCOUNT` and `AZURE_STORAGE_KEY` environment variables. Optionally, you can also set `AZURE_STORAGE_SAS_TOKEN` if you are using a Shared Access Signature (SAS) for authentication.

### Using Azure Blob Storage
You can use Azure Blob Storage by setting the `HBOX_STORAGE_CONN_STRING` to `azblob://my-container`.

### Local Azure Storage Emulator
If you want to use the local Azure Storage Emulator, you can set the `HBOX_STORAGE_CONN_STRING` to `azblob://my-container?protocol=http&domain=localhost:10001`. This will allow you to use the emulator for development and testing purposes.