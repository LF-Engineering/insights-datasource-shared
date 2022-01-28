###### Terraform FAQs


- whenever the [backend](https://github.com/LF-Engineering/insights-datasource-shared/blob/main/deploy/dev/main.tf#L8) is changed please use following commands locally to migrate the backend first

`terraform init -migrate-state` & check with `terraform plan` if the new configuration is correctly migrated.

- if the [backend](https://github.com/LF-Engineering/insights-datasource-shared/blob/main/deploy/dev/main.tf#L8) is not setup, initially initialize it locally using following command

`terraform init`