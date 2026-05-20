terraform {
  required_version = ">= 1.5.0"
  
  required_providers {
    huaweicloud = {
      source  = "huaweicloud/huaweicloud"
      version = ">= 1.56.0"
    }
    
    random = {
      source  = "hashicorp/random"
      version = ">= 3.5.0"
    }
    
    tls = {
      source  = "hashicorp/tls"
      version = ">= 4.0.0"
    }
  }
  
  backend "local" {
    path = "terraform.tfstate"
  }
}