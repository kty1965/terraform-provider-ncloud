---
subcategory: "Load Balancer"
---


# Resource: ncloud_lb_listener

Provides a Load Balancer Listener resource.

~> **NOTE:** This resource only supports VPC environment.

## Example Usage
```hcl
resource "ncloud_lb" "test" {
  # ...
}

resource "ncloud_lb_target_group" "test" {
  # ...
}

resource "ncloud_lb_listener" "test" {
  load_balancer_no = ncloud_lb.test.load_balancer_no
  protocol = "HTTP"
  port = 80
  target_group_no = ncloud_lb_target_group.test.target_group_no
}
```

## Argument Reference

The following arguments are supported:

* `load_balancer_no` - (Required) The ID of the load balancer.
* `target_group_no` - (Required) The ID of the target group.
* `port` - (Required) The port on which the load balancer is listening. Valid from `1` to `65534`.
* `protocol` - (Required) The protocol type for the listener. The types of protocols available are limited by the type of load balancer. `APPLICATION` Load Balancer Accepted values: `HTTP` | `HTTPS`, `NETWORK` Load Balancer Accepted values : `TCP`, `UDP`, `NETWORK_PROXY` Load Balancer Accepted values : `TCP` | `TLS`. 
* `tls_min_version_type` - (Optional) The TLS minimum supported version type code. Valid only if the listener protocol type is `HTTPS` or `TLS`. Accepted values : `TLSV10`(TLSv1.0) | `TLSV11`(TLSv1.1) | `TLSV12`(TLSv1.2). Default: `TLSV10`.
* `use_http2` - (Optional) Whether to use HTTP/2 protocol. Valid only if the listener protocol type is `HTTPS`. Accepted values : `true`, `false`. Default: `false`.
* `ssl_certificate_no` - (Optional) The ID of the SSL certificate. If the listener protocol type is `HTTPS` or `TLS`, an SSL certificate must be set.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of listener.
* `listener_no` - The ID of listener (It is the same result as id).
* `rule_no_list` - The list of listener rule number.

## Import

### `terraform import` command

* Load Balancer Listener can be imported using the `load_balancer_no`:`listener_no`. For example:

```console
$ terraform import ncloud_lb_listener.rsc_name 17019658:12345
```

### `import` block

* In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Load Balancer Listener using the `load_balancer_no`:`listener_no`. For example:

```terraform
import {
  to = ncloud_lb_listener.rsc_name
  id = "17019658:12345"
}
```
