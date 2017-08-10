# creates an ec2 instance for the purposes of being
# a jenkins agent with a specific profile
#   python agent-creator.py <profilename> <keypair-name> <boxname> <contactemail>
# example
#   python agent-creator.py default COPS-master jenkins-slave-ecs-control russell.endicott@ge.com

import boto3
import json
import sys

profilename = sys.argv[1]
keyname = sys.argv[2]
boxname = sys.argv[3]
contactemail = sys.argv[4]
s = boto3.session.Session(profile_name=profilename)
c = s.client('ec2')
tags = [
                {
                    'Key': 'Name',
                    'Value': boxname
                },
                {
                    'Key': 'env',
                    'Value': 'prod'
                },
                {
                    'Key': 'uai',
                    'Value': 'UAI2008328'
                },
                {
                    'Key': 'contact',
                    'Value': contactemail
                }
            ]

response = c.run_instances(
    DryRun=False,
    ImageId='ami-9c2c0ce7',
    MinCount=1,
    MaxCount=1,
    KeyName=keyname,
    SecurityGroupIds=[
        'sg-1c5b7a63',
    ],
    InstanceType='t2.medium',
    Placement={
        'AvailabilityZone': 'us-east-1c',
        'Tenancy': 'default'
    },
    BlockDeviceMappings=[
        {
            'DeviceName': "/dev/sda1",
            'Ebs': {
                'VolumeSize': 50,
                'DeleteOnTermination': False,
                'VolumeType': 'standard',
            }
        },
    ],
    Monitoring={
        'Enabled': False
    },
    SubnetId='subnet-4f507d62',
    DisableApiTermination=False,
    InstanceInitiatedShutdownBehavior='stop',
    IamInstanceProfile={
        'Arn': 'arn:aws:iam::188894168332:instance-profile/ecsInstanceRoleY'
    },
    EbsOptimized=False

)

print(response)

instanceId = response.get('Instances')[0].get('InstanceId')

response = c.create_tags(
    DryRun = False,
    Resources = [
        instanceId,
    ],
    Tags = tags
)
