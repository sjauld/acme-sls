module "acme_sls" {
  source = "../"

  certificates = {
    "acme-sls.viostream.xyz" = ["acme-sls.viostream.xyz", "subdomain2.acme-sls.viostream.xyz"]
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
  count = length(module.acme_sls.buckets)

  zone_id = data.aws_route53_zone.viostream_xyz.id

  name    = module.acme_sls.buckets[count.index].bucket
  type    = "CNAME"
  ttl     = 300
  records = [module.acme_sls.buckets[count.index].bucket_domain_name]
}
