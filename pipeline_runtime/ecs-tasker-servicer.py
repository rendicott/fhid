#!/usr/bin/env python
# Creates a new task from a given task definition json and starts on
# all instances in the given cluster name
# USAGE:
#  python ecs-tasker.py <task_definition_json_filename> <cluster_name>
# EXAMPLE:
#  python ecs-tasker.py ecs-task-stage.json cops-cluster

import boto3
import json
import sys
import time
from pprint import pprint

fname = sys.argv[1]
cluster_name = sys.argv[2]
service_name = 'fhid-service-stage'
target_group_arn = 'arn:aws:elasticloadbalancing:us-east-1:188894168332:targetgroup/tg-fhid-stage/597e8afd93568c05'
container_name = 'fhid-stage'
container_port = 80
desired_count = 2
sleeptime = 20
role_arn = 'arn:aws:iam::188894168332:role/ecrAccess'

fmt_logs_uri = "https://us-east-1.console.aws.amazon.com/cloudwatch/home?region=us-east-1#logEventViewer:group=awslogs-ecs;stream=awslogs-fhid-stage/fhid-stage/{0}"

with open(fname,'rb') as f:
    task = json.load(f)

s = boto3.session.Session()
c = s.client('ecs', region_name='us-east-1')

def create_service(task_definition):
    tries = 0
    max_tries = 3
    print("Attempt %d of %d..." % (tries, max_tries))
    while 1:
        if tries > max_tries:
            print("Max tries exceeded, exiting with failure....")
            sys.exit(1)
        try:
            response = c.create_service(
                cluster=cluster_name,
                serviceName=service_name,
                taskDefinition=task_definition,
                loadBalancers=[
                    {
                        'targetGroupArn': target_group_arn,
                        'containerName': container_name,
                        'containerPort': container_port
                    },
                ],
                desiredCount=desired_count,
                role=role_arn,
                deploymentConfiguration={
                    'maximumPercent': 200,
                    'minimumHealthyPercent': 100
                },
                placementConstraints=[],
                placementStrategy=[{
                    "field": "memory",
                    "type": "binpack"
                }
                ]
            )

            print response
            break
        except Exception as e:
            print("Exception creating service: '%s'" % str(e))
            tries += 1
            print("Sleeping...")
            time.sleep(5)



container_instances = c.list_container_instances(cluster=cluster_name).get('containerInstanceArns')

response = c.register_task_definition(containerDefinitions=task.get('containerDefinitions'),
                                      networkMode=task.get('networkMode'),
                                      taskRoleArn=task.get('taskRoleArn'),
                                      family=task.get('family'))

definition = response.get('taskDefinition').get('taskDefinitionArn')


def task_tester():
    retries = 1
    max_retries = 5
    tasks = []
    while 1:
        print("Attempt %d of %d..." % (retries, max_retries))
        if retries > max_retries:
            print("Too many task start failures")
            sys.exit(1)
        tasker = c.start_task(taskDefinition=definition,
                            cluster=cluster_name,
                            containerInstances=[container_instances[0]]) # max of 10 instances

        print("Sleeping %d seconds to wait for tasks to start..." % sleeptime)
        time.sleep(sleeptime)
        print("Number of tasks started: %d" % len(tasker.get('tasks')))
        if len(tasker.get('failures')) > 0:
            print("Number of failed tasks: %d" % len(tasker.get('failures')))
            for failure in tasker.get('failures'):
                print(failure)
                if failure.get('reason') == "RESOURCE:MEMORY":
                    retries += 1
        else:
            break

    all_tasks = c.list_tasks(cluster=cluster_name)
    all_tasks_arns = all_tasks.get('taskArns')
    for task_arn in c.describe_tasks(cluster=cluster_name, tasks=all_tasks_arns).get('tasks'):
        if task_arn.get('taskDefinitionArn') == definition:
            tasks.append(task_arn.get('taskArn'))

    status = c.describe_tasks(cluster=cluster_name,
                            tasks=tasks)
    return tasks

tasks = task_tester()
# check on status of tasks and exit with failure if 
# containers don't stay running
count = 0
maxCount = 10
FAILED = False
RUNNING = False
runningCount = 0
task_definition_arn = ""
task_arn = ""
while 1:
    count += 1
    status = c.describe_tasks(cluster=cluster_name,
                          tasks=tasks)
    for task in status.get('tasks'):

        if task.get('lastStatus') == "STOPPED":
            print("CONTAINER FAILED:")
            pprint(status)
            FAILED = True
            try:
                guid = task.get('taskArn').split('/')[-1]
                print("LOGS URL: %s" % fmt_logs_uri.format(guid))
            except:
                pass
            break
        if task.get('lastStatus') == "PENDING":
            print("Task still PENDING...sleeping")
        else:
            pprint(status)
            task_definition_arn = task.get('taskDefinitionArn')
            task_arn = task.get("taskArn")
            RUNNING = True
            break
    if count > maxCount:
        print("Too many iterations, exiting status failed.")
        FAILED = True
    if FAILED:
        break
    if RUNNING:
        runningCount += 1
    if runningCount > 3:
        create_service(task_definition_arn)
        c.stop_task(cluster=cluster_name,
                    task=task_arn,
                    reason="Temporary task for pipeline build")
        break
    time.sleep(5)

if FAILED:
    sys.exit(1)
else:
    sys.exit(0)



