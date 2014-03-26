# HTTP proxy apparatus for Gilliam

This simple apparatus implements a HTTP proxy that makes it possible
to talk directly to other services running in Gilliam without
integrating with the service registry.

# Getting Started

Add the following to you `gilliam.yml` file to have the apparatus
install next time you build your app:

    apparatuses:
      - http://github.com/gilliam/apparatus-http-proxy/releases/<ver>.tgz

The `HTTP_PROXY` environment variable will be set and point to proxy.

# Building

...

