# The Reading List CLI

Use `gcloud auth application-default login` to generate credentials.

Alternatively, you can use the `-creds` flag that points to the path of a Google service account JSON key file.

If running locally, use `-insecure` and `-fakeID` to inject user identity.

## Usage

```
Usage of ./cli:
  -creds string
    	the path of the service account credentials file. if empty, uses Google Application Default Credentials.
  -delete
    	delete this URL from the list (requires -mode update)
  -fakeID string
    	for local development - a user ID to inject into the request
  -host string
    	the host of the reading list server (default "localhost:8081")
  -insecure
    	use an insecure connection
  -limit int
    	limit for the number of links to return when listing links (default 20)
  -mode string
    	(list|update) (default "list")
  -url string
    	the URL to add or delete
```
