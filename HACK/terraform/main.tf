provider "aws" {
  region = "eu-west-1"
  version = "= 1.10.0"
}

terraform {
  backend "s3" {
    bucket = "aws.all.kaptain.terraform"
    key = "terraform.tfstate"
    region = "eu-west-1"
  }
}
