provider "aws" {
  region = var.eg_aws_region
  secret_key = var.aws_secret_key
  access_key = var.aws_access_key
}

terraform {
  backend "s3" {
    bucket         = "insights-v2-dev"
    key            = "terraform/connector-ecs-tasks/terraform.tfstate"
    region         = "us-east-2" # this cant be replaced with the variable
    encrypt        = true
    kms_key_id     = "alias/terraform-bucket-key"
  }
}

resource "aws_kms_key" "terraform-bucket-key" {
  description             = "This key is used to encrypt bucket objects"
  deletion_window_in_days = 10
  enable_key_rotation     = true
}

resource "aws_kms_alias" "key-alias" {
  name          = "alias/terraform-bucket-key"
  target_key_id = aws_kms_key.terraform-bucket-key.key_id
}

resource "aws_s3_bucket" "terraform-state" {
  bucket = "insights-v2-dev"

  tags = {
    Name        = "Insights V2 Dev"
    Environment = "dev"
  }
}

resource "aws_s3_bucket_acl" "terraform-state-acl" {
  bucket = aws_s3_bucket.terraform-state.id
  acl    = "private"
}

resource "aws_s3_bucket_server_side_encryption_configuration" "terraform-state-encryption-configuration" {
  bucket = aws_s3_bucket.terraform-state.id

    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = aws_kms_key.terraform-bucket-key.arn
        sse_algorithm     = "aws:kms"
      }
    }
}

resource "aws_s3_bucket_public_access_block" "block" {
  bucket = aws_s3_bucket.terraform-state.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

/* ECS cluster */
resource "aws_ecs_cluster" "insights-ecs-cluster" {
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
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-git:latest"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration: {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "insights-ecs-git",
          "awslogs-region": var.eg_aws_region,
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
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-jira:latest"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration: {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "insights-connector-jira-logs",
          "awslogs-region": var.eg_aws_region,
          "awslogs-create-group": "true",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ])

}

/* ECS confluence connector task definition */
resource "aws_ecs_task_definition" "insights-connector-confluence-task" {
  family = "insights-connector-confluence-task"
  requires_compatibilities = ["FARGATE"]
  network_mode = "awsvpc"
  cpu = "256"
  memory = "512"
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn = aws_iam_role.ecs_task_role.arn
  container_definitions = jsonencode([
    {
      name      = "insights-connector-confluence"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-confluence:latest"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration: {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "insights-ecs-confluence",
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
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-gerrit:latest"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration: {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "insights-connector-gerrit-task",
          "awslogs-region": var.eg_aws_region,
          "awslogs-create-group": "true",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ])

}


/* ECS bugzilla connector task definition */
resource "aws_ecs_task_definition" "insights-connector-bugzilla-task" {
  family = "insights-connector-bugzilla-task"
  requires_compatibilities = ["FARGATE"]
  network_mode = "awsvpc"
  cpu = "256"
  memory = "512"
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn = aws_iam_role.ecs_task_role.arn
  container_definitions = jsonencode([
    {
      name      = "insights-connector-bugzilla"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-bugzilla:latest"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration: {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "insights-ecs-bugzilla",
          "awslogs-region": "us-east-2",
          "awslogs-create-group": "true",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ])

}


/* ECS github task definitions */
resource "aws_ecs_task_definition" "insights-connector-github-task" {
  family = "insights-connector-github-task"
  requires_compatibilities = ["FARGATE"]
  network_mode = "awsvpc"
  cpu = "256"
  memory = "512"
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn = aws_iam_role.ecs_task_role.arn
  container_definitions = jsonencode([
    {
      name      = "insights-connector-github"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-github:latest"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration: {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "insights-connector-github-task",
          "awslogs-region": var.eg_aws_region,
          "awslogs-create-group": "true",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ])

}

/* ECS bugzillarest connector task definition */
resource "aws_ecs_task_definition" "insights-connector-bugzillarest-task" {
  family = "insights-connector-bugzillarest-task"
  requires_compatibilities = ["FARGATE"]
  network_mode = "awsvpc"
  cpu = "256"
  memory = "512"
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn = aws_iam_role.ecs_task_role.arn
  container_definitions = jsonencode([
    {
      name      = "insights-connector-bugzillarest"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-bugzillarest:latest"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration: {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "insights-ecs-bugzillarest",
          "awslogs-region": "us-east-2",
          "awslogs-create-group": "true",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ])

}

/* ECS dockerhub connector task definition */
resource "aws_ecs_task_definition" "insights-connector-dockerhub-task" {
  family = "insights-connector-dockerhub-task"
  requires_compatibilities = ["FARGATE"]
  network_mode = "awsvpc"
  cpu = "256"
  memory = "512"
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn = aws_iam_role.ecs_task_role.arn
  container_definitions = jsonencode([
    {
      name      = "insights-connector-dockerhub"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-dockerhub:latest"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration: {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "insights-ecs-dockerhub",
          "awslogs-region": "us-east-2",
          "awslogs-create-group": "true",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ])

}

/* ECS jenkins connector task definition */
resource "aws_ecs_task_definition" "insights-connector-jenkins-task" {
  family = "insights-connector-jenkins-task"
  requires_compatibilities = ["FARGATE"]
  network_mode = "awsvpc"
  cpu = "256"
  memory = "512"
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn = aws_iam_role.ecs_task_role.arn
  container_definitions = jsonencode([
    {
      name      = "insights-connector-jenkins"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-jenkins:latest"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration: {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "insights-ecs-jenkins",
          "awslogs-region": "us-east-2",
          "awslogs-create-group": "true",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ])

}

/* ECS circleci connector task definition */
resource "aws_ecs_task_definition" "insights-connector-circleci-task" {
  family = "insights-connector-circleci-task"
  requires_compatibilities = ["FARGATE"]
  network_mode = "awsvpc"
  cpu = "256"
  memory = "512"
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn = aws_iam_role.ecs_task_role.arn
  container_definitions = jsonencode([
    {
      name      = "insights-connector-circleci"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-circleci:latest"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration: {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "insights-ecs-circleci",
          "awslogs-region": "us-east-2",
          "awslogs-create-group": "true",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ])

}

/* ECS rocketchat connector task definition */
resource "aws_ecs_task_definition" "insights-connector-rocketchat-task" {
  family = "insights-connector-rocketchat-task"
  requires_compatibilities = ["FARGATE"]
  network_mode = "awsvpc"
  cpu = "256"
  memory = "512"
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn = aws_iam_role.ecs_task_role.arn
  container_definitions = jsonencode([
    {
      name      = "insights-connector-rocketchat"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-rocketchat:latest"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration: {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "insights-ecs-rocketchat",
          "awslogs-region": "us-east-2",
          "awslogs-create-group": "true",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ])

}

/* ECS pipermail connector task definition */
resource "aws_ecs_task_definition" "insights-connector-pipermail-task" {
  family = "insights-connector-pipermail-task"
  requires_compatibilities = ["FARGATE"]
  network_mode = "awsvpc"
  cpu = "256"
  memory = "512"
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn = aws_iam_role.ecs_task_role.arn
  container_definitions = jsonencode([
    {
      name      = "insights-connector-pipermail"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-pipermail:latest"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration: {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "insights-ecs-pipermail",
          "awslogs-region": "us-east-2",
          "awslogs-create-group": "true",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ])

}

/* ECS groupsio connector task definition */
resource "aws_ecs_task_definition" "insights-connector-groupsio-task" {
  family = "insights-connector-groupsio-task"
  requires_compatibilities = ["FARGATE"]
  network_mode = "awsvpc"
  cpu = "256"
  memory = "1024"
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn = aws_iam_role.ecs_task_role.arn
  container_definitions = jsonencode([
    {
      name      = "insights-connector-groupsio"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-groupsio:latest"
      cpu       = 256
      memory    = 1024
      essential = true
      logConfiguration: {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "insights-ecs-groupsio",
          "awslogs-region": "us-east-2",
          "awslogs-create-group": "true",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ])

}

/* ECS googlegroups connector task definition */
resource "aws_ecs_task_definition" "insights-connector-googlegroups-task" {
  family = "insights-connector-googlegroups-task"
  requires_compatibilities = ["FARGATE"]
  network_mode = "awsvpc"
  cpu = "256"
  memory = "512"
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn = aws_iam_role.ecs_task_role.arn
  container_definitions = jsonencode([
    {
      name      = "insights-connector-googlegroups"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-googlegroups:latest"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration: {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "insights-ecs-googlegroups",
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
  cluster         = aws_ecs_cluster.insights-ecs-cluster.id
  task_definition = aws_ecs_task_definition.insights-connector-git-task.arn
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"
  network_configuration {
    security_groups = [aws_security_group.security_group.id]
    subnets = [aws_subnet.main.id]
    assign_public_ip = true
  }

}


resource "aws_ecs_service" "github" {
  name            = "insights-github"
  cluster         = aws_ecs_cluster.insights-ecs-cluster.id
  task_definition = aws_ecs_task_definition.insights-connector-github-task.arn
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"
  network_configuration {
    security_groups = [aws_security_group.security_group.id]
    subnets = [aws_subnet.main.id]
    assign_public_ip = true
  }
}

resource "aws_ecs_service" "jira" {
  name            = "insights-jira"
  cluster         = aws_ecs_cluster.insights-ecs-cluster.id
  task_definition = aws_ecs_task_definition.insights-connector-jira-task.arn
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"
  network_configuration {
    security_groups = [aws_security_group.security_group.id]
    subnets = [aws_subnet.main.id]
    assign_public_ip = true
  }
}

resource "aws_ecs_service" "gerrit" {
  name            = "insights-gerrit"
  cluster         = aws_ecs_cluster.insights-ecs-cluster.id
  task_definition = aws_ecs_task_definition.insights-connector-gerrit-task.arn
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"
  network_configuration {
    security_groups = [aws_security_group.security_group.id]
    subnets = [aws_subnet.main.id]
    assign_public_ip = true
  }
}

/* ecs bugzilla service */
resource "aws_ecs_service" "bugzilla" {
  name            = "insights-bugzilla"
  cluster         = aws_ecs_cluster.insights-ecs-cluster.id
  task_definition = aws_ecs_task_definition.insights-connector-bugzilla-task.arn
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"
  network_configuration {
    security_groups = [aws_security_group.security_group.id]
    subnets = [aws_subnet.main.id]
    assign_public_ip = true
  }

}

/* ecs bugzillarest service */
resource "aws_ecs_service" "bugzillarest" {
  name            = "insights-bugzillarest"
  cluster         = aws_ecs_cluster.insights-ecs-cluster.id
  task_definition = aws_ecs_task_definition.insights-connector-bugzillarest-task.arn
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"
  network_configuration {
    security_groups = [aws_security_group.security_group.id]
    subnets = [aws_subnet.main.id]
    assign_public_ip = true
  }

}

/* ecs dockerhub service */
resource "aws_ecs_service" "dockerhub" {
  name            = "insights-dockerhub"
  cluster         = aws_ecs_cluster.insights-ecs-cluster.id
  task_definition = aws_ecs_task_definition.insights-connector-dockerhub-task.arn
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"
  network_configuration {
    security_groups = [aws_security_group.security_group.id]
    subnets = [aws_subnet.main.id]
    assign_public_ip = true
  }

}

/* ecs circleci service */
resource "aws_ecs_service" "circleci" {
  name            = "insights-circleci"
  cluster         = aws_ecs_cluster.insights-ecs-cluster.id
  task_definition = aws_ecs_task_definition.insights-connector-circleci-task.arn
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"
  network_configuration {
    security_groups = [aws_security_group.security_group.id]
    subnets = [aws_subnet.main.id]
    assign_public_ip = true
  }

}

/* ecs confluence service */
resource "aws_ecs_service" "confluence" {
  name            = "insights-confluence"
  cluster         = aws_ecs_cluster.insights-ecs-cluster.id
  task_definition = aws_ecs_task_definition.insights-connector-confluence-task.arn
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"
  network_configuration {
    security_groups = [aws_security_group.security_group.id]
    subnets = [aws_subnet.main.id]
    assign_public_ip = true
  }

}

/* ecs rocketchat service */
resource "aws_ecs_service" "rocketchat" {
  name            = "insights-rocketchat"
  cluster         = aws_ecs_cluster.insights-ecs-cluster.id
  task_definition = aws_ecs_task_definition.insights-connector-rocketchat-task.arn
  launch_type                        = "FARGATE"
  scheduling_strategy                = "REPLICA"
  network_configuration {
    security_groups = [aws_security_group.security_group.id]
    subnets = [aws_subnet.main.id]
    assign_public_ip = true
  }

}

/* ecs jenkins service */
resource "aws_ecs_service" "jenkins" {
  name            = "insights-jenkins"
  cluster         = aws_ecs_cluster.insights-ecs-cluster.id
  task_definition = aws_ecs_task_definition.insights-connector-jenkins-task.arn
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
  name = "ecs-ta-role"

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
  name = "ecs-tas-role"

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

resource "aws_iam_role_policy_attachment" "task_role_ssm_policy_attachment" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMReadOnlyAccess"
}

resource "aws_iam_role_policy_attachment" "task_role_cloudwatch_policy_attachment" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess"
}

resource "aws_iam_role_policy_attachment" "task_execution_role_cloudwatch_policy_attachment" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess"
}

data "aws_iam_policy_document" "kms_use" {
  statement {
    sid = ""
    effect = "Allow"
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:DescribeKey",
    ]
    resources = [
      "arn:aws:kms:us-east-2:${var.eg_account_id}:key/f36a45d3-9bce-4f10-bedc-5a20c2ff807e"
    ]
  }
}

resource "aws_iam_policy" "kms_use" {
  name        = "kmsuse"
  description = "Policy to allow use of KMS Key"
  policy      = data.aws_iam_policy_document.kms_use.json
}
