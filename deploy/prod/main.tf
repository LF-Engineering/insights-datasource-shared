provider "aws" {
  region = var.eg_aws_region
  secret_key = var.aws_secret_key
  access_key = var.aws_access_key
}

/* ECS cluster */
resource "aws_ecs_cluster" "insights-git-cluster" {
  name = "insights-ecs-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

/* ECS task definitions */
resource "aws_ecs_task_definition" "insights-connector-git-task" {
  family = "insights-connector-git-task"
  requires_compatibilities = ["FARGATE"]
  network_mode = "awsvpc"
  cpu = "256"
  memory = "512"
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn = aws_iam_role.ecs_task_role.arn
  container_definitions = jsonencode([
    {
      name      = "insights-connector-git"
      image     = "395594542180.dkr.ecr.us-east-1.amazonaws.com/insights-connector-git:stable"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration: {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "insights-ecs-git",
          "awslogs-region": "us-east-2",
          "awslogs-create-group": "true",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ])

}

/* ECS task definitions */
resource "aws_ecs_task_definition" "insights-connector-jira-task" {
  family = "insights-connector-jira-task"
  requires_compatibilities = ["FARGATE"]
  network_mode = "awsvpc"
  cpu = "256"
  memory = "512"
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn = aws_iam_role.ecs_task_role.arn
  container_definitions = jsonencode([
    {
      name      = "insights-connector-jira"
      image     = "395594542180.dkr.ecr.us-east-1.amazonaws.com/insights-connector-jira:stable"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration: {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "insights-connector-jira-logs",
          "awslogs-region": "us-east-2",
          "awslogs-create-group": "true",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ])

}

/* ECS gerrit connector task definition */
resource "aws_ecs_task_definition" "insights-connector-gerrit-task" {
  family = "insights-connector-gerrit-task"
  requires_compatibilities = ["FARGATE"]
  network_mode = "awsvpc"
  cpu = "256"
  memory = "512"
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn = aws_iam_role.ecs_task_role.arn
  container_definitions = jsonencode([
    {
      name      = "insights-connector-gerrit"
      image     = "395594542180.dkr.ecr.us-east-l.amazonaws.com/insights-connector-gerrit:stable"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration: {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "insights-connector-gerrit-task",
          "awslogs-region": "us-east-2",
          "awslogs-create-group": "true",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ])

}


resource "aws_security_group" "security_group" {
  name        = "example-task-security-group"
  vpc_id      = aws_vpc.main.id

  ingress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}
resource "aws_internet_gateway" "gateway" {
  vpc_id = aws_vpc.main.id
}

resource "aws_default_route_table" "public" {
  default_route_table_id = aws_vpc.main.main_route_table_id
}

resource "aws_route" "public_internet_gateway" {
  route_table_id         = aws_default_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.gateway.id

  timeouts {
    create = "5m"
  }
}

resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.main.id
  route_table_id = aws_default_route_table.public.id
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "main" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
  map_public_ip_on_launch = true
  tags = {
    Name = "Main"
  }
}

/* ecs service */
resource "aws_ecs_service" "git" {
  name            = "insights-git"
  cluster         = aws_ecs_cluster.insights-git-cluster.id
  task_definition = aws_ecs_task_definition.insights-git-task.arn
  desired_count   = 1
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"
  network_configuration {
    security_groups = [aws_security_group.security_group.id]
    subnets = [aws_subnet.main.id]
    assign_public_ip = true
  }

}


/* iam roles */

resource "aws_iam_role" "ecs_task_execution_role" {
  name = "role-name"

  assume_role_policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Action": "sts:AssumeRole",
     "Principal": {
       "Service": [
          "ecs-tasks.amazonaws.com",
          "cloudwatch.amazonaws.com"
        ]
     },
     "Effect": "Allow",
     "Sid": ""
   }
 ]
}
EOF
}

resource "aws_iam_role" "ecs_task_role" {
  name = "role-name-task"

  assume_role_policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Action": "sts:AssumeRole",
     "Principal": {
       "Service": "ecs-tasks.amazonaws.com"
     },
     "Effect": "Allow",
     "Sid": ""
   }
 ]
}
EOF
}

/* policy attachments */
resource "aws_iam_role_policy_attachment" "ecs-task-execution-role-policy-attachment" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy_attachment" "task_role_s3_policy_attachment" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonS3FullAccess"
}

resource "aws_iam_role_policy_attachment" "task_role_cloudwatch_policy_attachment" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess"
}

resource "aws_iam_role_policy_attachment" "task_execution_role_cloudwatch_policy_attachment" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess"
}