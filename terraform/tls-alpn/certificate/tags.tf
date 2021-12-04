locals {
  tf_module      = "acme-sls-certificate"
  tf_module_repo = "github.com/sjauld/acme-sls"
}

locals {
  tf_path = "${lookup(var.tags, "tf_path", "root")}.${local.tf_module}"

  tags = merge(
    var.tags,
    {
      "tf_module"      = local.tf_module
      "tf_module_repo" = local.tf_module_repo
      "tf_path"        = local.tf_path
    },
  )
}
