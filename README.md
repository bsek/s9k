## How to build

Run `make` 

## Prerequisites

The application currently looks for container images in Github Container Registry and lambda functions in a S3 bucket. For this to work, two requirements needs to be fulfilled. 

1. An S3 bucket name must be provided in an environment variable named: `S3_DEPLOYMENT_BUCKET_NAME` and must be organized like this:

```
<bucket_name>/<function_name>/<version>.zip
```

2. It needs a Github PAT. During startup the application queries the gh cli tool for a valid token. It expects the gh cli tool to be installed and configured with a valid token.


