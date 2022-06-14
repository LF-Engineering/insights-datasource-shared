provider "aws" {
  region     = var.eg_aws_region
  secret_key = var.aws_secret_key
  access_key = var.aws_access_key
}

terraform {
  backend "s3" {
    bucket     = "insights-v2-terraform-state-prod"
    key        = "terraform/connector-ecs-tasks/terraform.tfstate"
    region     = "us-east-2" # this cant be replaced with the variable
    encrypt    = true
    kms_key_id = "alias/terraform-bucket-key"
  }
}

resource "aws_kms_key" "terraform-bucket-key" {
  description             = "This key is used to encrypt bucket data"
  deletion_window_in_days = 10
  enable_key_rotation     = true
}

resource "aws_kms_alias" "key-alias" {
  name          = "alias/terraform-bucket-key"
  target_key_id = aws_kms_key.terraform-bucket-key.key_id
}

resource "aws_s3_bucket" "terraform-state" {
  bucket = "insights-v2-terraform-state-prod"

  tags = {
    Name        = "Insights V2 terraform state Prod"
    Environment = "prod"
  }
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

/* iam roles */

resource "aws_iam_role" "ecs_task_execution_role" {
  name = "insights-ecs-task-execution-role"

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
  name = "insights-ecs-task-role"

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

/* ECS task definitions */
resource "aws_ecs_task_definition" "insights-connector-git-task" {
  family                   = "insights-connector-git-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "1024"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions    = jsonencode([
    {
      name      = "insights-connector-git"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-git:stable"
      cpu       = 256
      memory    = 1024
      essential = true
      logConfiguration : {
        "logDriver" : "awslogs",
        "options" : {
          "awslogs-group" : "insights-ecs-connector-git",
          "awslogs-region" : var.eg_aws_region,
          "awslogs-create-group" : "true",
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])

}

/* ECS task definitions */
resource "aws_ecs_task_definition" "insights-connector-jira-task" {
  family                   = "insights-connector-jira-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions    = jsonencode([
    {
      name      = "insights-connector-jira"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-jira:stable"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration : {
        "logDriver" : "awslogs",
        "options" : {
          "awslogs-group" : "insights-ecs-connector-jira",
          "awslogs-region" : var.eg_aws_region,
          "awslogs-create-group" : "true",
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])

}

/* ECS confluence connector task definition */
resource "aws_ecs_task_definition" "insights-connector-confluence-task" {
  family                   = "insights-connector-confluence-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "512"
  memory                   = "2048"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions    = jsonencode([
    {
      name      = "insights-connector-confluence"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-confluence:stable"
      cpu       = 512
      memory    = 2048
      essential = true
      logConfiguration : {
        "logDriver" : "awslogs",
        "options" : {
          "awslogs-group" : "insights-ecs-connector-confluence",
          "awslogs-region" : var.eg_aws_region,
          "awslogs-create-group" : "true",
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])

}

/* ECS gerrit connector task definition */
resource "aws_ecs_task_definition" "insights-connector-gerrit-task" {
  family                   = "insights-connector-gerrit-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions    = jsonencode([
    {
      name      = "insights-connector-gerrit"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-gerrit:stable"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration : {
        "logDriver" : "awslogs",
        "options" : {
          "awslogs-group" : "insights-ecs-connector-gerrit",
          "awslogs-region" : var.eg_aws_region,
          "awslogs-create-group" : "true",
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])

}

/* ECS bugzilla connector task definition */
resource "aws_ecs_task_definition" "insights-connector-bugzilla-task" {
  family                   = "insights-connector-bugzilla-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions    = jsonencode([
    {
      name      = "insights-connector-bugzilla"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-bugzilla:stable"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration : {
        "logDriver" : "awslogs",
        "options" : {
          "awslogs-group" : "insights-ecs-connector-bugzilla",
          "awslogs-region" : var.eg_aws_region,
          "awslogs-create-group" : "true",
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])

}

/* ECS github task definitions */
resource "aws_ecs_task_definition" "insights-connector-github-task" {
  family                   = "insights-connector-github-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "512"
  memory                   = "2048"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions    = jsonencode([
    {
      name      = "insights-connector-github"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-github:stable"
      cpu       = 512
      memory    = 2048
      essential = true
      logConfiguration : {
        "logDriver" : "awslogs",
        "options" : {
          "awslogs-group" : "insights-ecs-connector-github",
          "awslogs-region" : var.eg_aws_region,
          "awslogs-create-group" : "true",
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])

}

/* ECS bugzillarest connector task definition */
resource "aws_ecs_task_definition" "insights-connector-bugzillarest-task" {
  family                   = "insights-connector-bugzillarest-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions    = jsonencode([
    {
      name      = "insights-connector-bugzillarest"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-bugzillarest:stable"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration : {
        "logDriver" : "awslogs",
        "options" : {
          "awslogs-group" : "insights-ecs-connector-bugzillarest",
          "awslogs-region" : var.eg_aws_region,
          "awslogs-create-group" : "true",
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])

}

/* ECS dockerhub connector task definition */
resource "aws_ecs_task_definition" "insights-connector-dockerhub-task" {
  family                   = "insights-connector-dockerhub-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions    = jsonencode([
    {
      name      = "insights-connector-dockerhub"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-dockerhub:stable"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration : {
        "logDriver" : "awslogs",
        "options" : {
          "awslogs-group" : "insights-ecs-connector-dockerhub",
          "awslogs-region" : var.eg_aws_region,
          "awslogs-create-group" : "true",
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])

}

/* ECS jenkins connector task definition */
resource "aws_ecs_task_definition" "insights-connector-jenkins-task" {
  family                   = "insights-connector-jenkins-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions    = jsonencode([
    {
      name      = "insights-connector-jenkins"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-jenkins:stable"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration : {
        "logDriver" : "awslogs",
        "options" : {
          "awslogs-group" : "insights-ecs-connector-jenkins",
          "awslogs-region" : var.eg_aws_region,
          "awslogs-create-group" : "true",
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])

}

/* ECS circleci connector task definition */
resource "aws_ecs_task_definition" "insights-connector-circleci-task" {
  family                   = "insights-connector-circleci-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions    = jsonencode([
    {
      name      = "insights-connector-circleci"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-circleci:stable"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration : {
        "logDriver" : "awslogs",
        "options" : {
          "awslogs-group" : "insights-ecs-connector-circleci",
          "awslogs-region" : var.eg_aws_region,
          "awslogs-create-group" : "true",
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])

}

/* ECS rocketchat connector task definition */
resource "aws_ecs_task_definition" "insights-connector-rocketchat-task" {
  family                   = "insights-connector-rocketchat-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions    = jsonencode([
    {
      name      = "insights-connector-rocketchat"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-rocketchat:stable"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration : {
        "logDriver" : "awslogs",
        "options" : {
          "awslogs-group" : "insights-ecs-connector-rocketchat",
          "awslogs-region" : var.eg_aws_region,
          "awslogs-create-group" : "true",
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])

}

/* ECS pipermail connector task definition */
resource "aws_ecs_task_definition" "insights-connector-pipermail-task" {
  family                   = "insights-connector-pipermail-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "512"
  memory                   = "2048"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions    = jsonencode([
    {
      name      = "insights-connector-pipermail"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-pipermail:stable"
      cpu       = 512
      memory    = 2048
      essential = true
      logConfiguration : {
        "logDriver" : "awslogs",
        "options" : {
          "awslogs-group" : "insights-ecs-connector-pipermail",
          "awslogs-region" : var.eg_aws_region,
          "awslogs-create-group" : "true",
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])

}

/* ECS groupsio connector task definition */
resource "aws_ecs_task_definition" "insights-connector-groupsio-task" {
  family                   = "insights-connector-groupsio-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "512"
  memory                   = "2048"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions    = jsonencode([
    {
      name      = "insights-connector-groupsio"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-groupsio:stable"
      cpu       = 512
      memory    = 2048
      essential = true
      logConfiguration : {
        "logDriver" : "awslogs",
        "options" : {
          "awslogs-group" : "insights-ecs-connector-groupsio",
          "awslogs-region" : var.eg_aws_region,
          "awslogs-create-group" : "true",
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])

}

/* ECS googlegroups connector task definition */
resource "aws_ecs_task_definition" "insights-connector-googlegroups-task" {
  family                   = "insights-connector-googlegroups-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions    = jsonencode([
    {
      name      = "insights-connector-googlegroups"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-googlegroups:stable"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration : {
        "logDriver" : "awslogs",
        "options" : {
          "awslogs-group" : "insights-ecs-connector-googlegroups",
          "awslogs-region" : var.eg_aws_region,
          "awslogs-create-group" : "true",
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])
}

/* ECS githubstats connector task definition */
resource "aws_ecs_task_definition" "insights-connector-githubstats-task" {
  family                   = "insights-connector-githubstats-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions    = jsonencode([
    {
      name      = "insights-connector-githubstats"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/insights-connector-githubstats:stable"
      cpu       = 128
      memory    = 512
      essential = true
      logConfiguration : {
        "logDriver" : "awslogs",
        "options" : {
          "awslogs-group" : "insights-ecs-connector-githubstats",
          "awslogs-region" : var.eg_aws_region,
          "awslogs-create-group" : "true",
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])
}

/* ECS scheduler connector task definition */
resource "aws_ecs_task_definition" "insights-scheduler-task" {
  family                   = "insights-scheduler-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "512"
  memory                   = "1024"
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn
  container_definitions    = jsonencode([
    {
      name      = "insights-scheduler"
      image     = "${var.eg_account_id}.dkr.ecr.${var.eg_aws_region}.amazonaws.com/lfx-insights-scheduler:stable"
      cpu       = 128
      memory    = 512
      essential = true
      environment : [
        {
          "name" : "SCHEDULER_ENVIRONMENT",
          "value" : "prod"
        },
        {
          "name" : "BUGZILLA_DISABLED",
          "value" : "true"
        },
        {
          "name" : "BUGZILLAREST_DISABLED",
          "value" : "true"
        },
        {
          "name" : "DOCKERHUB_DISABLED",
          "value" : "true"
        },
        {
          "name" : "JIRA_DISABLED",
          "value" : "true"
        },
        {
          "name" : "CIRCLECI_DISABLED",
          "value" : "true"
        },
        {
          "name" : "JENKINS_DISABLED",
          "value" : "true"
        },
        {
          "name" : "ROCKETCHAT_DISABLED",
          "value" : "true"
        },
        {
          "name" : "GOOGLEGROUPS_DISABLED",
          "value" : "true"
        },
        {
          "name" : "PIPERMAIL_DISABLED",
          "value" : "true"
        },
        {
          "name" : "GROUPSIO_DISABLED",
          "value" : "true"
        },
        {
          "name" : "CONFLUENCE_DISABLED",
          "value" : "true"
        }
      ]
      secrets : [
        {
          name : "SCHEDULER_ES_CACHE_URL",
          valueFrom : "arn:aws:ssm:${var.eg_aws_region}:${var.eg_account_id}:parameter/insights/elastic_cache_url"
        },
        {
          name : "SCHEDULER_ES_LOG_URL",
          valueFrom : "arn:aws:ssm:${var.eg_aws_region}:${var.eg_account_id}:parameter/insights/elastic_log_url"
        },
        {
          name : "SCHEDULER_CONNECTOR_API",
          valueFrom : "arn:aws:ssm:${var.eg_aws_region}:${var.eg_account_id}:parameter/connectors_api_url"
        },
        {
          name : "SCHEDULER_CONN_STRING",
          valueFrom : "arn:aws:ssm:${var.eg_aws_region}:${var.eg_account_id}:parameter/insights/onboarding/postgresql/prod"
        },
        {
          name : "SCHEDULER_WEB_HOOK_URL",
          valueFrom : "arn:aws:ssm:${var.eg_aws_region}:${var.eg_account_id}:parameter/insights/slackwebhookurl"
        },
        {
          name : "SCHEDULER_AUTH_GRANT_TYPE",
          valueFrom : "arn:aws:ssm:${var.eg_aws_region}:${var.eg_account_id}:parameter/insights/auth0_grant_type"
        },
        {
          name : "SCHEDULER_AUTH_CLIENT_ID",
          valueFrom : "arn:aws:ssm:${var.eg_aws_region}:${var.eg_account_id}:parameter/insights/auth0_client_id"
        }, {
          name : "SCHEDULER_AUTH_CLIENT_SECRET",
          valueFrom : "arn:aws:ssm:${var.eg_aws_region}:${var.eg_account_id}:parameter/insights/auth0_client_secret"
        },
        {
          name : "SCHEDULER_AUTH_AUDIENCE",
          valueFrom : "arn:aws:ssm:${var.eg_aws_region}:${var.eg_account_id}:parameter/insights/auth0_audience"
        }, {
          name : "SCHEDULER_AUTH0_URL",
          valueFrom : "arn:aws:ssm:${var.eg_aws_region}:${var.eg_account_id}:parameter/insights/auth0_url"
        },
        {
          name : "SCHEDULER_GAP_URL",
          valueFrom : "arn:aws:ssm:${var.eg_aws_region}:${var.eg_account_id}:parameter/insights/elastic_gap_url"
        },
        {
          name : "SCHEDULER_CIRCLECI_TOKEN",
          valueFrom : "arn:aws:ssm:${var.eg_aws_region}:${var.eg_account_id}:parameter/insights/circleci_token"
        }
      ],
      logConfiguration : {
        "logDriver" : "awslogs",
        "options" : {
          "awslogs-group" : "insights-ecs-connector-scheduler",
          "awslogs-region" : var.eg_aws_region,
          "awslogs-create-group" : "true",
          "awslogs-stream-prefix" : "ecs"
        }
      }
    }
  ])
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "security_group" {
  name   = "example-task-security-group"
  vpc_id = aws_vpc.main.id

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

resource "aws_subnet" "main" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true
  tags                    = {
    Name = "Main"
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

/* ecs scheduler service */
resource "aws_ecs_service" "insights-scheduler" {
  name                = "insights-scheduler"
  cluster             = aws_ecs_cluster.insights-ecs-cluster.id
  task_definition     = aws_ecs_task_definition.insights-scheduler-task.arn
  desired_count       = 1
  launch_type         = "FARGATE"
  scheduling_strategy = "REPLICA"
  network_configuration {
    security_groups  = [aws_security_group.security_group.id]
    subnets          = [aws_subnet.main.id]
    assign_public_ip = true
  }
}

/* policy attachments */
resource "aws_iam_policy" "ssm_get_parameters_policy" {
  name        = "ssm-get-parameters"
  description = "A ssm get params policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ssm:GetParameters"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "task_role_ssm_get_parameters_policy_attachment" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = aws_iam_policy.ssm_get_parameters_policy.arn
}

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
    sid       = ""
    effect    = "Allow"
    actions   = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:DescribeKey",
    ]
    resources = [
      "arn:aws:kms:${var.eg_aws_region}:${var.eg_account_id}:key/0434cd25-c409-43af-8b69-8873fbf227f8"
    ]
  }
}

resource "aws_iam_policy" "kms_use" {
  name        = "kmsuse"
  description = "Policy allows using KMS keys"
  policy      = data.aws_iam_policy_document.kms_use.json
}

resource "aws_iam_role_policy_attachment" "test-attach" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.kms_use.arn
}
