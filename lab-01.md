# Lab 1 - Writing a Composition Function

In this lab, we will walk through a hands-on guide that demonstrates all the
steps required to write a simple [Composition
Function](https://docs.crossplane.io/latest/concepts/composition-functions/).

By the end of this lab, you will have a running Function that creates an S3
bucket for every named entry in a composite resource (XR). For example, the
following sample XR will create 3 separate S3 buckets:

```yaml
apiVersion: example.crossplane.io/v1
kind: XBuckets
metadata:
  name: example-buckets
spec:
  region: us-east-2
  names:
  - crossplane-functions-example-a
  - crossplane-functions-example-b
  - crossplane-functions-example-c
```

## Lab Content

The full content of this hands-on lab can be found in the Crossplane
documentation:

https://docs.crossplane.io/knowledge-base/guides/write-a-composition-function-in-go

Please work through this entire guide and think through each step carefully to
understand the programming model for Functions end to end. Having this full
example should help bring all the pieces together. We'll be right here to answer
questions and get you unblocked if you're stuck!