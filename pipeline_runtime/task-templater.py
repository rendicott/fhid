#!/usr/bin/env python
# generates an output json template for an ECS task definition
# created so a task definition could be created to spin up a 
# specific version of a container instead of just 'latest'
# USAGE:
#  python task-templater.py <user_template_file> <version_tag> <output_filename>
# SAMPLE:
#  python task-templater.py ecs-task-prod.json.tpl 2.0.1.49 ecs-task-prod.json
import sys
import jinja2

user_template = sys.argv[1]
version_tag = sys.argv[2]
target_filename = sys.argv[3]

templateLoader = jinja2.FileSystemLoader( searchpath="./")
templateEnv = jinja2.Environment( loader=templateLoader )
TEMPLATE_FILE = user_template
template = templateEnv.get_template( TEMPLATE_FILE )
outputText = template.render(versionTag=version_tag)

with open(target_filename,'wb') as f:
    f.write(outputText)