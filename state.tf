terraform {
  backend "s3" {
    profile = "default"
    region  = "us-east-1"
    key     = "stop-ec2-instances/terraform.tfstate"
    bucket  = "sbali-tfstate"
  }
}