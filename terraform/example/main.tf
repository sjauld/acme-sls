provider "aws" {
  region = "us-east-1"
}

module "acme_sls" {
  source = "../"

  certificates = {
    "example1.viostream.xyz" = ["example1.viostream.xyz", "blog.example1.viostream.xyz"],
    "example2.viostream.xyz" = ["example2.viostream.xyz", "blog.example2.viostream.xyz"],
  }

  tags = {
    "env" : "sand"
    "project" : "ACME-SLS"
  }
}

data "aws_route53_zone" "viostream_xyz" {
  name = "viostream.xyz"
}

# This only works because all the domains on all the certificates can be CNAMEd
# from the same R53 zone - for more complicated setups you'll have to do
# something more complicated
resource "aws_route53_record" "acme_sls_viostream_xyz" {
  count = length(module.acme_sls.cname_records)

  zone_id = data.aws_route53_zone.viostream_xyz.id

  name    = module.acme_sls.cname_records[count.index].name
  type    = "CNAME"
  ttl     = 300
  records = [module.acme_sls.cname_records[count.index].record]
}
