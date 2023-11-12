terraform {
  required_providers {
    arvan = {
      source = "arvancloud/arvan"
      version = "0.6.4"
    }
  }
}
variable "ApiKey" {  
    type = string  
    default = "apikey 0027674a-a12b-5c83-adca-a4f8c7672711"  
    sensitive = true 
} 
provider "arvan" {  
    api_key = var.ApiKey
} 
variable "abrak-name" { 

 type = string 

 default = "kuber1" 

} 

variable "region" { 

 type = string 

 default = "ir-thr-c1" # Forogh Datacenter 

} 

resource "arvan_iaas_abrak" "kuber1" { 

 region = var.region 

 flavor = "g1-4-8-0" 

 name  = var.abrak-name 

 image { 

  type = "distributions" 

  name = "debian/11" 

 } 

 disk_size = 25 

} 

data "arvan_iaas_abrak" "get_abrak_id" { 

 depends_on = [ 

  arvan_iaas_abrak.kuber1

 ] 


 region = var.region 

 name  = var.abrak-name 

} 

output "details-abrak-1" { 

 value = data.arvan_iaas_abrak.get_abrak_id 

} 