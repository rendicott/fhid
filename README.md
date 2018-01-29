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

Run `go test ./... -v` from the root of the repo.


# Usage

## Post

Submitting an entry would look something like this flow. The first step would be to post the results of an image build:
```
curl -XPOST https://images.company.com/v1.0/images -d '{
"Version":"1.2.4",
"BaseOS":"Arch",
"BuildNotes":{
	"BuildLog": ["line one","line two"],
	"OutputAmis": [
		{"AmiID": "ami-54321","AmiRegion":"us-west-1", 
		 "AmiTags":[{"Name":"test","Value":"test"}],
		 "AmiSharedTo": ["1234567","7654321"]}
	]
},
"ReleaseNotes":{}
}'
```

Which would return:

```
{"Success": "True", "Data": "e9373eb2-b17f-4344-a933-4db2d358c020"}
```

Then once the resulting output AMI has been tested you would then release it to the world and then update the record like so:
```
curl -XUPDATE https://images.company.com/v1.0/images?ImageID=e9373eb2-b17f-4344-a933-4db2d358c020 -d '{

"ReleaseNotes":{
	"ReleaseNote": "Pushing out a thing to do that dingy",
	"Amis": [
		{"AmiID": "ami-54321","AmiRegion":"us-west-1", 
		 "AmiTags":[{"Name":"test","Value":"test"}],
		 "AmiSharedTo": ["1234567","7654321","67183674","10239485"]},
		{"AmiID": "ami-54322","AmiRegion":"us-east-1", 
		 "AmiTags":[{"Name":"test","Value":"test"}],
		 "AmiSharedTo": ["1234567","7654321","67183674","10239485"]}
	]
}
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
		"ImageID": "e9373eb2-b17f-4344-a933-4db2d358c020",
		"Version": "1.2.4",
		"BaseOS": "Arch",
		"BuildNotes": {
            "BuildLog": ["line one","line two"],
            "OutputAmis": [
                {"AmiID": "ami-54321","AmiRegion":"us-west-1", 
                "AmiTags":[{"Name":"test","Value":"test"}],
                "AmiSharedTo": ["1234567","7654321"]}
            ]},
		"CreateDate": "2017-08-22 22:40:04"
	}]
}
```

## GET

You can also just do a targeted `GET` if you include the `?ImageID=<id>` image ID in the query string parameter. 

Example:

```
https://images.company.com/v1.0/images?ImageID=30095350-dd02-4200-bf12-894f409a653f
```

## PATCH

You can update the `ReleaseNotes` section on an existing entry by using the `PATCH` method on the image endpoint and including the image ID you'd like to update. 

```
curl -XPATCH https://images.company.com/v1.0/image?ImageID=30095350-dd02-4200-bf12-894f409a653f -d '{
"ReleaseNotes":{
	"ReleaseNote": "Pushing out a thing to do that dingy",
	"Amis": [
		{"AmiID": "ami-54321","AmiRegion":"us-west-1", 
		 "AmiTags":[{"Name":"test","Value":"test"}],
		 "AmiSharedTo": ["1234567","7654321","67183674","10239485"]},
		{"AmiID": "ami-54322","AmiRegion":"us-east-1", 
		 "AmiTags":[{"Name":"test","Value":"test"}],
		 "AmiSharedTo": ["1234567","7654321","67183674","10239485"]}
	]
}
}'
```

Currently only the `ReleaseNotes` attribute of the object can be updated. 

## supported queries

| function name | supported values | description |
|---------------|------------------|-------------|
| `StringMatch` | regex patterns   | compiles the regex pattern and uses it to search the given field |
