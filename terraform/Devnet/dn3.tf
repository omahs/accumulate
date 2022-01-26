data "aws_ecs_task_definition" "dn-3" {
  task_definition = "${aws_ecs_task_definition.dn-3.family}"
}

resource "aws_ecs_task_definition" "dn-3" {
  family = "dn-3"
  container_definitions = <<DEFINITION
[

   {
      "name": "dn-3",
      "image": "registry.gitlab.com/accumulatenetwork/accumulate/accumulated:develop",
      "essential": true,
      "portMappings": [{"containerPort": 3000}],
      "memory": 512,
      "cpu": 256,
      "logConfiguration": {
         "logDriver": "awslogs",
         "options": {
            "awslogs-group": "accumulate-devnet-log",
            "awslogs-region": "us-east-1",
            "awslogs-stream-prefix": "ecs"
                }
            },
      "command": ["run", "-w", "/mnt/efs/node/dn/Node3"],
      "mountPoints": [
           {
               "containerPath": "/mnt/efs/node",
               "sourceVolume": "efs_temp",
               "readOnly": false
           }
       ]
       
    }
]
DEFINITION
  requires_compatibilities = ["FARGATE"] # Stating that we are using ECS Fargate
  network_mode             = "awsvpc"    # Using awsvpc as our network mode as this is a  Fargate requirement
  memory                   = 512   # Specifying the memory our container requires
  cpu                      = 256      # Specifying the CPU our container requires
  volume {
    name     = "efs_temp"
    efs_volume_configuration {
      file_system_id = "${aws_efs_file_system.devnet.id}"
      root_directory = "/mnt/efs/node"
      transit_encryption      = "ENABLED"
      transit_encryption_port = 2999
      authorization_config {
        access_point_id = aws_efs_access_point.devnet.id
        iam = "ENABLED"
      }
    }
  }

  execution_role_arn       = "${aws_iam_role.ecsTaskExecutionRole_1.arn}"
  task_role_arn            = "${aws_iam_role.ecsTaskExecutionRole_1.arn}"

}

  resource "aws_ecs_service" "dn-3" {
  name            = "dn-3"              
  cluster         = "${aws_ecs_cluster.dev_cluster.id}"
  task_definition = "${aws_ecs_task_definition.dn-3.arn}"
  launch_type     = "FARGATE"
  desired_count   = 1
  platform_version = "1.4.0"


  load_balancer {
    target_group_arn = "${aws_alb_target_group.dev_target.arn}" # Reference our target group
    container_name   = "${aws_ecs_task_definition.dn-3.family}"
    container_port   = 3000
  }

  service_registries {
      registry_arn = aws_service_discovery_service.devnet-dn3.arn
      container_name = "dn-3"
  }

  network_configuration {
    subnets          = ["${aws_subnet.dev_pubsub_a.id}","${aws_subnet.dev_pubsub_b.id}"]
    assign_public_ip = true # Provide our containers with public IPs
    security_groups  = ["${aws_security_group.devnet.id}"]
      }
    depends_on = [aws_alb_listener.dev_listener]
}

resource "aws_service_discovery_service" "devnet-dn3" {
  name = "dn-3"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.devnet.id

    dns_records {
      ttl  = 10
      type = "A"
    }


    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }
}
