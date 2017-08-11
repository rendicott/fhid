# fhid
## Fixham Harbour Image Depot
_Created for an image tracking project dubbed 'Bob the Builder'_

Provides a REST interface to store and retrieve entries to a Redis database for the purposes of tracking images such as Amazon AMI's, OpenStack base images, and other virtualization base templates.

The problem with most base images is that they're usually a black box and nobody ever seems to know what's on an image. This is an attempt to provide a location and a structure with with to track these images. 

# dev usage
fire up a local redis server then run `go run main.go -c dev-config.json -loglevel debug`
You can post json to the `/images` handler
and then `GET` to the `/images` handler with a query like `/images?ImageId=d07d13d9-b666-46d6-986f-a57c4ee8e971`


# Usage

## Post

Submitting an entry would look something like this:
```
curl -XPOST http://localhost:8090 -d '{
"Version": "1.2.3.142",
"BaseOS": "Centos6.6",
"ReleaseNotes": "Stuff"
}'
```
Any other fields will just be ignored. 

## Query
Searching for an image with certain properties can be done by `POST` to the `/query` handler.

Query format looks like this
```
{
	"<fieldname>": {"<function>": "<regex_pattern>|<function_name>"}
}
```

For example, to search for an image with `Ubuntu` in the `BaseOS` field:
```
{
	"BaseOS": {"StringMatch": ".*Ubuntu.*"}
}
```


## supported queries

| function name | supported values | description |
| `StringMatch` | regex patterns   | compiles the regex pattern and uses it to search the given field |