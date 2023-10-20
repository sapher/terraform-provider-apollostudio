terraform {
  required_providers {
    apollo = {
      source = "hashicorp.com/sapher/apollostudio"
    }
  }
}

provider "apollo" {
  org_id = "your-org-id"
}
