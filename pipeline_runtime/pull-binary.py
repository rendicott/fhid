#!/usr/bin/env python
# Pull down latest build and installs locally using provided credentials

# USAGE:
#  python pull-binary.py <bucketname> <path+filename> <aws_access_key_id> <aws_secret_access_key>

# EXAMPLE:
#  python pull-binary.py "my-unique-bucket-name" "/sub/folder/something.tar.gz" "AKIALFOIENLANVVJQ" "89shty123M9xoHqG/ElGCHbi3ScHyZT4ls+szOVp"

import boto3
import tarfile
import sys

bucketname = sys.argv[1]
pathAndFname = sys.argv[2]
accesskey = sys.argv[3]
secretkey = sys.argv[4]

fnames = []
fname = pathAndFname.split('/')[-1]
fnames.append(fname)

# use IAM creds for 502722899 user in DPCO account
client = boto3.client(
    's3',
    aws_access_key_id=accesskey,
    aws_secret_access_key=secretkey
)

for f in fnames:
    client.download_file(
        bucketname,
        pathAndFname,
        './' + f
    )

    tar = tarfile.open(fname, "r:gz")
    tar.extractall()
    tar.close()