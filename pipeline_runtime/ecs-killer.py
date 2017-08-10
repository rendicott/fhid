#!/usr/bin/env python
# Stop, delete, and kill all tasks from an ecs service
# USAGE:
#  python ecs-killer.py <cluster_name> <service_name>
# SAMPLE:
#  python ecs-killer.py api-read api-read-service-stage

import boto3
import sys
from pprint import pprint

cluster_name = sys.argv[1]
service_name = sys.argv[2]

s = boto3.session.Session()
c = s.client('ecs', region_name='us-east-1')


sarns = c.list_services(cluster=cluster_name).get('serviceArns')
dasarn = ""
dadef = ""

for sarn in sarns:
    response = c.describe_services(
        cluster=cluster_name,
        services=[
            sarn,
        ]
    )
    for k in response.get('services'):
        if k.get('serviceName') == service_name:
            dasarn = sarn
            dadef = k.get('taskDefinition')

print("Working on destroying sarn: %s" % dasarn)

if dasarn != "":
    try:
        service = c.describe_services(
                cluster=cluster_name,
                services=[
                    dasarn,
                ])

        #pprint(service)

        c.update_service(cluster=cluster_name, service=dasarn, desiredCount=0)

        c.delete_service(cluster=cluster_name, service=dasarn)


        print("Working on destroying tasks with taskdef: %s" % dadef)

        tokill = []
        for task in c.describe_tasks(cluster=cluster_name,
                                    tasks=c.list_tasks(cluster=cluster_name).get('taskArns')).get('tasks'):
            if task.get('taskDefinitionArn') == dadef:
                tokill.append(task.get('taskArn'))

        for tarn in tokill:
            print("Working on killing taskArn: %s" % tarn)
            c.stop_task(cluster=cluster_name, task=tarn, reason="Pipeline temp testing done")
    except Exception as arr:
        print("Exception killing service and/or tasks: " + str(arr))
else:
    print("No service '%s' found." % service_name)

