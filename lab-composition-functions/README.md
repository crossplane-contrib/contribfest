# Composition Functions (xfn) for Crossplane

The composition function tutorial presented in KubeCon Amsterdam 2023 as part of
[Crossplane ContribFest](https://kccnceu2023.sched.com/event/1Hzcf).

This tutorial is a step-by-step guide to create composition functions that can
do things that would not be possible with the standard Composition such as:
* Creating resources conditionally,
* Generating a random string only once,
* Initialize infrastructure like creating a database schema or setting timezone
  for a database,
* Manipulating JSON values such as IAM policies in smarter ways.

We will write all composition functions in Golang.

## Index

* [Prerequisites](01-prerequisites.md)
* A no-op function to serve as boilerplate: [xfn-noop](02-xfn-noop.md)
* A function that assigns a randomly calculated value as default: [xfn-random](03-xfn-random.md)
* A function that creates N number of resources as specified by the user.
