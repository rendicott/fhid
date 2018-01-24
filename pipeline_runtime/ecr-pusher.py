#!/usr/bin/env python
# Uploads image to ECR
# Usage:
#  python ecr-pusher.py <ecr_repo_name> <local_image_name> <version_tag>
# SAMPLE:
#  python ecr-pusher.py api-read-stage api-read v2.1.3

import boto3
import docker
import os
import sys

repo_name = sys.argv[1]
image_name = sys.argv[2]
version = sys.argv[3]

region='us-east-1'
BOTO = boto3.session.Session()
DOCK = docker.from_env()
ECR = BOTO.client('ecr', region_name=region)

def add_image(ECR, DOCK):
    """Push local image to ECS"""
    for r in ECR.describe_repositories()['repositories']:
        if r.get('repositoryName') == repo_name:
            target_repo = r

    target_image = image_name + ":latest"

    #Log in
    print "\nLogging in...."
    output = os.popen("/home/jenkins/.local/bin/aws ecr get-login --region us-east-1 --no-include-email").read()
    os.system(output)

    #Docker Tag
    repo_info_latest = str(target_repo['repositoryUri']+':latest')
    repo_info = str(target_repo['repositoryUri']+":"+version)
    os.system('docker tag %s %s' % (target_image, repo_info))
    os.system('docker tag %s %s' % (target_image, repo_info_latest))
    print '\nTagged Docker Image as %s' % (repo_info)
    print '\nTagged Docker Image as %s' % (repo_info_latest)

    #Push Image
    print "\nPushing %s as %s..." % (target_image, repo_info)
    os.system('docker push %s' % (repo_info))
    print "\nPushing %s as %s..." % (target_image, repo_info_latest)
    os.system('docker push %s' % (repo_info_latest))

    #Untag Images
    print "\nUntagging image..."
    os.system('docker rmi %s' % (repo_info))
    os.system('docker rmi %s' % (repo_info_latest))

def main():
    add_image(ECR,DOCK)

if __name__ == '__main__':
    main()