# Lab 2 - Writing your own Function

Now that we've walked through creating an example function in the [previous
lab](./lab-01.md), let's get creative and write our own Function to do anything
we want!

## Examples

Let's look at some example Function ideas to get the creative juices flowing.
There are probably two high level scenarios you can think of for which writing a
Function may be useful to your platform and organization's needs:

1. **Specific:** An abstraction and a set of composed resources regularly used
   by your organization that you need to expose to your developers
    * Example: a Database abstraction that conditionally creates a managed
    PostgreSQL or MySQL database service in a few different clouds depending on
    input from the developer
1. **Generic:** A general composition scenario that can be reused by many
   Crossplane adopters in their Function pipelines
    * Example: a function that automatically detects when composed resources in
    a pipeline should be marked as `Ready`.  See
    https://github.com/crossplane-contrib/function-auto-ready for inspiration