# fhid
Fixham Harbour Image Depot
REST interface to the Bob the Builder Image repository


# post

Submitting an entry would look something like this:
```
curl -XPOST https://images.cloudpod.apps.ge.com/aws/ -d '{
"Version": "1.2.3.142",
"BaseOS": "Centos6.6",
"ReleaseNotes": {"Type": "Commit", "Content":"https://github.build.ge.com/212601587/aouta/commit/c549e33828cc3f78a300e1626d9fb055449d1d0f"}
}'
```