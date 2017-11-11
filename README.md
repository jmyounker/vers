Vers: A Version Tool
====================

Generating version numbers for software is a recurring release engineering
problem.  By providing a single tool we can eliminate a tiny corner of the
release engineering space.

Vers uses a combination of a configuration file, the command line operations,
and the revision control status to produce a unique version number.

Usage
-----
Generating a version file for your poject is the first step.

```
> vers -f version.json init
> cat version.json
{
  "data": {},
  "branches": [
    {
      "branch": ".*",
      "version": "{branch}.{commit-counter}"
    }
  ],
  "data-file": [
    "branch",
    "commit-counter",
    "version",
    "commit-hash",
    "commit-hash-short"
  ]
}
```

The default file above will generate something like
the following:

```
> vers -f version.json show 
master.15
```

Once you've edited the file to your satisfaction it
gets checked into your source tree.

```
> git add version.json
> git commit -m "add initial version file"
```

After committing you the generated version is different because
you have added one more commit to your branch.

```
> vers  -f version.json show
master.16
```

When you build a project you'll want to reference version information
from your code at runtime, or to communicate details about the build
environment to other steps.  You can do this with with a version data file.

```
> vers -f version.json data-file 
{
  "branch": "master",
  "commit-counter": 32,
  "commit-hash": "ab873...498fe",
  "commit-hash-short": "ab8743",
  "version": "master.32"
}
```

`Vers` can write the ata file directly to a file.

```
vers -f version.json data-file -o data-version.json
````

Semantic versioning
-------------------

`Vers` provides tooling for semantic versioning.

You add semantic versioning by editing the config file,
or by creating one initially with the `semvar` template

```
> vers -f version.json init --template semvar
> cat version.json
{
  "data": {
    "major": 0,
    "minor": 0,
    "release": 1
  },
  "branches": [
    {
      "branch": ".*",
      "version": "{major}.{minor}.{release}"
    }
  ],
  "data-file": [
    "branch",
    "commit-counter",
    "major",
    "minor",
    "release",
    "version",
    "commit-hash",
    "commit-hash-short"
  ]
}
```

Now we get a semantic version.
```
> vers -f version.json show 
0.0.1
```

`Vers` stores the version information in the version file's `data` section,
and it includes tools for updating the file.

```
> vers -f version.json bump-release 
cat version.json
{
  "data": {
    "major": 0,
    "minor": 0,
    "release": 2
  },
  "branches": [
    {
      "branch": ".*",
      "version": "{major}.{minor}.{release}"
    }
  ],
  "data-file": [
    "branch",
    "commit-counter",
    "major",
    "minor",
    "release",
    "version",
    "commit-hash",
    "commit-hash-short"
  ]
}
> vers -f version.json show
0.0.2
```

For the change to become permanent, you'll have to check in the modified
version file.

Similarly there are `bump-minor` and `bump-major` commands.

```
> vers -f version.json bump-minor
> cat version.json
{
  "data": {
    "major": 0,
    "minor": 1,
    "release": 0
  },
  ...
}
> vers -f version.json show
0.1.0

> vers -f version.json bump-major
> cat version.json
{
  "data": {
    "major": 1,
    "minor": 0,
    "release": 0
  },
  ...
}
> vers -f version.json show
1.0.0
```

As a final node, fields that look like numbers can be zero prefixed
to a fixed width.


Per-branch Formatting
---------------------

Often development organizations what to produce differently formatted
versions depending upon the phase of development/release cycle. This
usually maps to branch names, which in turn are often specially formatted.

`Vers` supports multiple version formats based on branch name pattern.

Production releases which come from master might use straight semantic
versioning, while development version might include branch and commit
counter information. The following example demonstrates this:

```
> cat version.json
{
  ...
  "branches": [
    {
      "branch": "master",
      "version": "{major}.{minor}.{release}"
    },
    {
      "branch": ".*",
      "version": "{major}.{minor}.{release}.{branch}.{commit-counter}"
    }
  ],
  ...
}

> git checkout master
> vers -f version.json show
1.0.1
> git checkout fb-parse-fields
> vers -f version.json
1.0.1.fb-parse-fields.16
```

The `branches` entries are processed from top to bottom, and the first
stanza with a `branch`  expression which matches is chosen.  The pattern
must be a complete match for the entire branch name, and the patterns
are go regular expressions.

Numeric Formatting
------------------

String expansions support fixed numeric field widths with leading
zero fill.  The expansion `{release}` expands the value `"5"` to `5`.
The expansion `{release:03d}` expands to `005`.

This lets you produce verions such as `1.3.02b004`.


Additional Information
----------------------

You can include arbitrary information in a version format or the
data file.

```
{
   ...
   "branches" : [
     {
       "branch": ".*",
       "format": "{major}.{minor}.{release}.b{build-id}"
     }
   ],
   ...
}
```

`Vers` accepts the data from the `-X` option.
```
> vers -f version.json show -X build-id=42
1.0.0.b42
```

The new information is mandatory when building a version.

```
> vers -f version.json show
could not expand build-id
```

The new information won't show up in the data file though.

```
{
  "branch": "master",
  "commit-counter": "16",
  "commit-hash": "d5a6d5e364dadb0c39751b9d07b2ca4d3d9eb834",
  "commit-hash-short": "d5a6d5e",
  "major": "1",
  "minor": "0",
  "release": "0",
  "version": "1.0.0.b42"
}
```
You'll have to add it to the data section yourself.

```
> vi version.json
...
> cat version.json
{
  ...
  "data-file": [
    "branch",
    "commit-counter",
    "build-id",
    "major",
    "minor",
    "release",
    "version",
    "commit-hash",
    "commit-hash-short"
  ]
}
```

And then the data file shows your the new `build-id` field.

```
> vers -f version.json data-file 
{
  "branch": "master",
  "build-id": "42",
  "commit-counter": "16",
  "commit-hash": "d5a6d5e364dadb0c39751b9d07b2ca4d3d9eb834",
  "commit-hash-short": "d5a6d5e",
  "major": "1",
  "minor": "0",
  "release": "0",
  "version": "1.0.0.b42"
}
```

Environment Variables
---------------------

`Vers` accepts parameter values from the environment also.  It
will check for the name as specified in file, and also for the
name in upper case with `-` changed to `-`.

The following example uses the preceding section's version pattern.

```
> export BUILD_ID=17
> vers -f version.json show 
1.0.0.b17
```

Overriding Parameter Values
---------------------------

You can override any parameter, even `branch`. This grants a great deal of
flexibility and facilitates rapid testing.  Using the variable `commit-counter`
as an example:

1. Command line through the `-X` option.
1. The environment variable `commit-counter`.
1. The environment variable `COMMIT_COUNTER`,
1. The parameter `commmit-counter` from the config file's `data` section.
1. The built-in functions.



