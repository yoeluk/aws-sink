## AWS Sink

AWS Sink is a `traefik` plugin that enable us to define a route to put data in S3 or dynamodb (not yet implemented) when `traefik` is deployed in ECS.

### Example configuration

`traefik.yml`
``` 
providers:
  ecs:
    exposedByDefault: false

experimental:
  plugins:
    aws-sink:
      moduleName: github.com/yoeluk/aws-sink
      version: v0.1.2
```
`ecs task labels` (only showing the most pertinent ecs task docker labels)
``` 
dockerLabels = {
    "traefik.enable" : "true"
    "traefik.http.routers.awssink.service" : "noop@internal"
    "traefik.http.routers.awssink.rule" : "Host(`awssink.myhostexample.io`)"
    "traefik.http.routers.awssink.middlewares" : "awssink"
    "traefik.http.middlewares.awsskink.plugin.aws-sink.sinkType" : "s3"
    "traefik.http.middlewares.awsskink.plugin.aws-sink.bucket" : "devjam-yoel"
    "traefik.http.middlewares.awsskink.plugin.aws-sink.region" : "us-west-2"
    "traefik.http.middlewares.awsskink.plugin.aws-sink.prefix" : "data"
}
```