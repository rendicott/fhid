# fhid
Fixham Harbour Image Depot
REST interface to the Bob the Builder Image repository

# dev usage
fire up a local redis server then run `go run main.go -c dev-config.json -loglevel debug`
You can post json to the `/images` handler
and then `GET` to the `/images` handler with a query like `/images?ImageId=d07d13d9-b666-46d6-986f-a57c4ee8e971`

# post

Submitting an entry would look something like this:
```
curl -XPOST https://images.cloudpod.apps.ge.com/aws/ -d '{
"Version": "1.2.3.142",
"BaseOS": "Centos6.6",
"ReleaseNotes": "Stuff"
}'
```
Any other fields will just be ignored. 