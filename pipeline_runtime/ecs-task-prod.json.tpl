{
    "networkMode": "bridge",
    "taskRoleArn": "arn:aws:iam::188894168332:role/ecrAccess",
    "containerDefinitions": [
        {
            "volumesFrom": [],
            "memory": 512,
            "portMappings": [
                {
                    "hostPort": 0,
                    "containerPort": 8090,
                    "protocol": "tcp"
                }
            ],
            "logConfiguration": {
                "logDriver": "awslogs",
                "options": {
                    "awslogs-group": "awslogs-ecs",
                    "awslogs-region": "us-east-1",
                    "awslogs-stream-prefix": "awslogs-fhid-prod"
                }
            },
            "name": "fhid-prod",
            "image": "188894168332.dkr.ecr.us-east-1.amazonaws.com/fhid-prod:{{ versionTag }}",
            "command": ["./fhid","-c","config.json","-loglevel","info"],
            "cpu": 0,
            "ulimits": [
                {
                    "name": "nofile",
                    "softLimit": 10000,
                    "hardLimit": 10000
                }]
        }
    ],
    "placementConstraints": [],
    "volumes": [],
    "family": "fhid-prod"
}