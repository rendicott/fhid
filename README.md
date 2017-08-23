# fhid
## Fixham Harbour Image Depot
_Created for an image tracking project dubbed 'Bob the Builder'_

Provides a REST interface to store and retrieve entries to a Redis database for the purposes of tracking images such as Amazon AMI's, OpenStack base images, and other virtualization base templates.

The problem with most base images is that they're usually a black box and nobody ever seems to know what's on an image. This is an attempt to provide a location and a structure with with to track these images. 

# dev usage
fire up a local redis server then run `go run main.go -c dev-config.json -loglevel debug`
You can post json to the `/images` handler
and then `GET` to the `/images` handler with a query like `/images?ImageId=d07d13d9-b666-46d6-986f-a57c4ee8e971`

_Testing_

To test make sure you have a local Redis server running on `127.0.0.1:6379` and then run `go test ./... -v` from the root of the repo.


# Usage

## Post

Submitting an entry would look something like this:
```
curl -XPOST https://images.company.com/v1.0/images -d '{
"Version": "3.4.5",
"BaseOS": "Mint14.04",
"ReleaseNotes": "The quick brown fox jumps over the lazy dog."
}'

curl -XPOST https://images.company.com/v1.0/images -d '{
"Version": "3.4.6",
"BaseOS": "Mint14.04",
"ReleaseNotes": "The quick brown fox jumps over the lazy dog."
}'

curl -XPOST https://images.company.com/v1.0/images -d '{
"Version": "3.4.7",
"BaseOS": "Mint14.04",
"ReleaseNotes": "The quick brown fox jumps over the lazy dog."
}'
```
Any other fields will just be ignored. 

## Query
Searching for an image with certain properties can be done by `POST` to the `/query` handler.

Query format looks like this
```
{
	"<fieldname>": {"<function>": "<regex_pattern>"}
}
```

For example, to search for an image with `Ubuntu` in the `BaseOS` field:
```
{
	"BaseOS": {"StringMatch": ".*Ubuntu.*"}
}
```

So to search for the three image entries we posted above would look like this:
```
curl -XPOST https://images.company.com/v1.0/query -d '{
"Version": {"StringMatch": "3.4.*"}
}'
```

Would return results:
```
{
	"Results": [{
		"ImageID": "5ea85df1-7c81-4061-aee2-613e97aa4b66",
		"Version": "3.4.5",
		"BaseOS": "Mint14.04",
		"ReleaseNotes": "The quick brown fox jumps over the lazy dog.",
		"CreateDate": "2017-08-22 22:40:04"
	}, {
		"ImageID": "89aeecc1-9072-4b74-a3f5-8177ecf018ef",
		"Version": "3.4.6",
		"BaseOS": "Mint14.04",
		"ReleaseNotes": "The quick brown fox jumps over the lazy dog.",
		"CreateDate": "2017-08-22 22:40:48"
	}, {
		"ImageID": "d896cb26-206e-48de-9467-ae0855b75710",
		"Version": "3.4.7",
		"BaseOS": "Mint14.04",
		"ReleaseNotes": "The quick brown fox jumps over the lazy dog.",
		"CreateDate": "2017-08-22 22:40:54"
	}]
}
```



## supported queries

| function name | supported values | description |
|---------------|------------------|-------------|
| `StringMatch` | regex patterns   | compiles the regex pattern and uses it to search the given field |
