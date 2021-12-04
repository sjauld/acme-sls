# lambda-http

This package is designed as a scheduled lambda, triggered by Cloudwatch Events.
It will kick off an HTTP-01 challenge with Let's Encrypt.

However, due to the inability to route traffic from port 80 to a Lambda, the
server part of the HTTP-01 challenge does not work. Therefore, you should either
use the lambda-tls package, or implement the HTTP-01 server somewhere else.
