<img src="assets/logo.svg" height=200 style="margin-bottom: 20px">

# Introduction
Ain is a terminal HTTP API client. It's an alternative to postman, paw or insomnia.

![Show and tell](assets/show-and-tell.gif?raw=true)

* Flexible organization of API:s using files and folders.
* Use shell-scripts and executables for common tasks.
* Put things that change in environment variables or .env-files.
* Handles url-encoding.
* Share the resulting [curl](https://curl.se/), [wget](https://www.gnu.org/software/wget/) or [httpie](https://httpie.io/) command-line.
* Pipe the API output for further processing.
* Tries hard to be helpful when there are errors.

Ain was built to enable scripting of input and further processing of output via pipes. It targets users who work with many API:s using a simple file format. It uses [curl](https://curl.se/), [wget](https://www.gnu.org/software/wget/) or [httpie](https://httpie.io/) to make the actual calls.

⭐ Please leave a star if you find it useful! ⭐

# Table of contents
<!-- npx doctoc --github --notitle --maxlevel=2 -->
<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Pre-requisites](#pre-requisites)
- [Installation](#installation)
  - [If you have go installed](#if-you-have-go-installed)
  - [Via homebrew](#via-homebrew)
  - [Via scoop](#via-scoop)
  - [Via the AUR (Arch Linux)](#via-the-aur-arch-linux)
  - [Download binaries yourself](#download-binaries-yourself)
- [Quick start](#quick-start)
- [Important concepts](#important-concepts)
- [Template files](#template-files)
- [Running ain](#running-ain)
- [Supported sections](#supported-sections)
  - [[Host]](#host)
  - [[Query]](#query)
  - [[Headers]](#headers)
  - [[Method]](#method)
  - [[Body]](#body)
  - [[Config]](#config)
  - [[Backend]](#backend)
  - [[BackendOptions]](#backendoptions)
- [Variables](#variables)
- [Executables](#executables)
- [Fatals](#fatals)
- [Quoting](#quoting)
- [Escaping](#escaping)
- [URL-encoding](#url-encoding)
- [Sharing is caring](#sharing-is-caring)
- [Handling line endings](#handling-line-endings)
- [Troubleshooting](#troubleshooting)
- [Ain in a bigger context](#ain-in-a-bigger-context)
- [Contributing](#contributing)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Pre-requisites
You need [curl](https://curl.se/), [wget](https://www.gnu.org/software/wget/) or [httpie](https://httpie.io/) installed and available on your `$PATH`. To test this run `ain -b`. This will generate a basic starter template listing what backends are available on your system in the [[Backend]](#backend) section. It will select one and leave the others commented out.

You can also check manually what backends you have installed by opening a shell and type `curl`, `wget` or `http` (add the suffix .exe to those commands if you're on windows). Any output from the command means it's installed.

On linux or mac one of the three is likely to already be installed. The others are available in your package manager or [homebrew](https://brew.sh).

If you're on windows curl.exe is installed if it's windows 10 build 17063 or higher. Otherwise you can get the binaries via [scoop](https://scoop.sh), [chocolatey](https://chocolatey.org/) or download them yourself. Ain uses curl.exe and cannot use the curl cmd-let powershell builtin.

# Installation

## If you have go installed
You need go 1.13 or higher. Using `go install`:
```
go install github.com/jonaslu/ain/cmd/ain@latest
```

## Via homebrew
Using the package-manager [homebrew](https://brew.sh)
```
brew install ain
```

## Via scoop
Using the windows package-manager [scoop](https://scoop.sh)
```
scoop bucket add jonaslu_tools https://github.com/jonaslu/scoop-tools.git
scoop install ain
```

## Via the AUR (Arch Linux)
From arch linux [AUR](https://aur.archlinux.org/) using [yay](https://github.com/Jguer/yay)
```
yay -S ain-bin
```

## Download binaries yourself
Install it so it's available on your `$PATH`:
[https://github.com/jonaslu/ain/releases](https://github.com/jonaslu/ain/releases)

# Quick start
Ain comes with a built in basic template that you can use as a starting point. Ain checks what backends (that's [curl](https://curl.se/), [wget](https://www.gnu.org/software/wget/) or [httpie](https://httpie.io/)) are available on your system and inserts them into the [[Backend]](#backend) section of the generated template. One will be selected and the rest commented out so the template is runnable directly.

Run:
```
ain -b basic-template.ain
```

The command above will output a starter-template to the file `basic-template.ain`.
The basic template calls the / GET http endpoint on localhost with the `Content-Type: application/json`.

To run the template specify a `PORT` variable:
```
ain basic-template.ain --vars PORT=8080
```

See the help for all options ain supports: `ain -h`

# Important concepts
* Templates: Files containing what, how and where to make the API call. By convention has the file suffix `.ain`.
* Sections: Label in a file grouping the API parameters.
* Variables: Things that vary as inputs in a template file.
* Executables: Enables using the output of a command in a template file.
* Backends: The thing that makes the API call ([curl](https://curl.se/), [wget](https://www.gnu.org/software/wget/) or [httpie](https://httpie.io/)).
* Fatals: Error in parsing the template files (it's your fault).

# Template files
Ain assembles data in template files to build the API-call. Ain parses the data following labels called [sections](#supported-sections) in each template file. Here's a full example:
```
[Host]           # The URL. Appends across files. Mandatory
http://localhost:${PORT}/api/blog/post

[Query]          # Query parameters. Appends across files
id=2e79870c-6504-4ac6-a2b7-01da7a6532f1

[Headers]        # Headers for the API-call. Appends across files
Authorization: Bearer $(./get-jwt-token.sh)
Content-Type: application/json

[Method]         # HTTP method. Overwrites across files
POST

[Body]           # Body for the API-call. Overwrites across files
{
  "title": "Reaping death",
  "content": "There is a place beyond the dreamworlds past the womb of night."
}

[Config]         # Ain specific config. Overwrites across files
Timeout=10

[Backend]        # How to make the API-call. Overwrites across files. Mandatory
curl

[BackendOptions] # Options to the selected backends. Appends across files
-sS              # Comments are ignored.
```
The template files can be named anything but some unique ending-convention such as .ain is recommended so you can [find](https://man7.org/linux/man-pages/man1/find.1.html) them easily.

Ain understands eight [Sections] with each of the sections described in details [below](#supported-sections). The data in sections either appends or overwrites across template files passed to ain.

Anything after a pound sign (#) is a comment and will be ignored.

# Running ain
`ain [OPTIONS] <template.ain> [--vars VAR=VALUE ...]` 

Ain accepts one or more template-file(s) as a mandatory argument. As sections appends or overwrite you can organize API-calls into hierarchical structures with increasing specificity using files and folders.

You can find examples of this in the [examples](https://github.com/jonaslu/ain/tree/main/examples) folder.

Adding an exclamation-mark (!) at the end of a template file name makes ain open the file in your `$VISUAL` or `$EDITOR` editor. If none is set it falls back to vim in that order. Once opened you edit the template file for this run only.

Example:
```
ain templates/get-blog-post.ain!     # Lets you edit the get-blog-post.ain for this run
```

Ain waits for the editor command to exit. Any terminal editor such as vim, emacs, nano etc will be fine. If your editor forks (as [vscode](https://code.visualstudio.com/) does by default) check if there's a flag stopping it from forking. To stop vscode from forking use the `--wait` [flag](https://code.visualstudio.com/docs/editor/command-line#_core-cli-options):

```
export EDITOR="code --wait"
```

If ain is connected to a pipe it will read template file names from the pipe. This enables you to use [find](https://man7.org/linux/man-pages/man1/find.1.html) and a selector such as [fzf](https://github.com/junegunn/fzf) to keep track of the template-files:
```
$> find . -name *.ain | fzf -m | ain
```

Template file names specified on the command line are read before names from a pipe. This means that `echo create-blog-post.ain | ain base.ain` is the same as `ain base.ain create-blog-post.ain`.

Ain functions as bash when it comes to file names: if they contain white-space the name must be quoted.

When making the call ain mimics how data is returned by the backend. After printing any internal errors of it's own, ain echoes back output from the backend: first the standard error (stderr) and then the standard out (stdout). It then returns the exit code from the backend command as it's own unless there are error specific to ain in which it returns status 1.

# Supported sections
Sections are case-insensitive and whitespace ignored but by convention uses CamelCase and are left indented. A section cannot be defined twice in a file. A section ends where the next begins or the file ends.

See [escaping](#escaping) If you need a literal supported section heading on a new line.

## [Host]
Contains the URL to the API. This section appends lines from one template file to the next. This feature allows you to specify a base-url in one file (e g `base.ain`) as such: `http://localhost:3000` and in the next template file specify the endpoint path (e g `login.ain`): `/api/auth/login`.

It's recommended that you use the [[Query]](#Query) section below for query-parameters as it handles joining with delimiters and trimming whitespace. You can however put raw query-parameters in the [Host] section too.

Any query-parameters added in the [[Query]](#Query) section are appended last to the URL. The whole URL is properly [url-encoded](#url-encoding) before passed to the backend. The [Host] section must combine to one and only one valid URL. Multiple URLs is not supported.

Ain performs no validation on the url (as backends differ on what a valid url looks like). If your call fails use `ain -p` as mentioned in [troubleshooting](#troubleshooting) to see what the run command looks like.

The [Host] section is mandatory and appends across template files.

## [Query]
All lines in the [Query] section is appended to the URL after it has been assembled. This means that you can specify query-parameters that apply to many endpoints in one file instead of having to include the same parameter in all endpoints.

An example is if an `API_KEY=<secret>` query-parameter applies to several endpoints. You can define this in a base-file and simply have the specific endpoint URL and possible extra query-parameters in their own file.

Example - `base.ain`:
```
[Host]
http://localhost:8080/api

[Query]
API_KEY=a922be9f-1aaf-47ef-b70b-b400a3aa386e
```

`get-post.ain`
```
[Host]
/blog/post

[Query]
id=1
```

This will result in the url:
```
http://localhost:8080/api/blog/post?API_KEY=a922be9f-1aaf-47ef-b70b-b400a3aa386e&id=1
```

The whitespace in a query key / value is only significant within the string.

This means that `page=3` and `page = 3` will become the same query parameter and `page = the next one` will become `page=the+next+one` when processed. If you need actual spaces between the equal-sign and the key / value strings you need to encode it yourself: e g `page+=+3` or put
that key-value in the [[Host]](#Host) section where space is significant.

Each line under the [Query] section is appended with a delimiter. Ain defaults to the query-string delimiter `&`. See the [[Config]](#Config) section for setting a custom delimiter.

All query-parameters are properly url-encoded. See [url-encoding](#url-encoding).

The [Query] section appends across template files.

## [Headers]
Headers to include in the API call.

Example:
```
[Headers]
Authorization: Bearer 888e90f2-319f-40a0-b422-d78bb95f229e
Content-Type: application/json
```

The [Headers] section appends across template files.

## [Method]
Http method (e g GET, POST, PATCH). If omitted the backend default is used (GET in both curl, wget and httpie).

Example:
```
[Method]
POST
```

The [Method] section is overridden by latter template files.

## [Body]
If the API call needs a body (as in the POST or PATCH http methods) the content of this section is passed as a file to the backend with formatting retained. Ain uses files to pass the [Body] contents because white-space may be important (e g yaml) and this section tends to be long.

The file is removed after the API call unless you pass the `-l` (as in leave) flag. Ain places the file in the $TMPDIR directory (usually `/tmp` on your box). You can override this in your shell by explicitly setting the `$TMPDIR` environment variable.

Passing the print command `-p` (as in print) flag will cause ain to write out the file named ain-body<random-digits> in the directory where ain is invoked and leave the file after completion. Leaving the body file makes the printed command shareable and runnable.

The [Body] section removes any leading and trailing whitespace lines, but keeps empty newlines between the first and last non-empty line.

Example:
```
[Body]

{
  "some": "json",  # ain removes comments

  "more": "jayson"
}

```

Is passed as this in the temp-file:
```
{

  "some": "json",

  "more": "jayson"
}
```

The [Body] section overwrites across template files.

## [Config]
This section contains config for ain. All config parameters are case-insensitive and any whitespace is ignored. Parameters for backends themselves are passed via the [[BackendOptions]](#BackendOptions) section.

Full config example:
```
[Config]
Timeout=3
QueryDelim=;
```

The [Config] sections overwrites across template files.

### Timeout
Config format: `Timeout=<timeout in seconds>`

The timeout is enforced during the whole execution of ain (both running executables and the actual API call). If omitted defaults to no timeout. This is the only section where [executables](#executables) cannot be used, since the timeout needs to be known before the executables are invoked.

### Query delimiter
Config format: `QueryDelim=<text>`

This is the delimiter used when concatenating the lines under the [[Query]](#Query) section. It can be any text that does not contain a space including the empty string.

Defaults to (`&`).

## [Backend]
The [Backend] specifies what command should be used to run the actual API call.

Valid options are [curl](https://curl.se/), [wget](https://www.gnu.org/software/wget/) or [httpie](https://httpie.io/).

Example:
```
[Backend]
curl
```

The [Backend] section is mandatory and overwrites across template files.

## [BackendOptions]
Backend specific options that are passed on to the backend command invocation.

Example:
```
[Backend]
curl

[BackendOptions]
-sS   # Makes curl disable its progress bar in a pipe
```

The [BackendOptions] section appends across template files.

# Variables
Variables lets you specify things that vary such as ports, item ids etc. Ain supports variables via environment variables. Anything inside `${}` in a template is replaced with the value found in the environment. Example `${NODE_ENV}`. Environment variables can be set in your shell in various ways, or via the `--vars VAR1=value1 VAR2=value2` syntax passed after all template file names.

This will set the variable values in ain:s environment (and available via inheritance in any `$(commands)` spawned from the template [executables](#executables)). Variables set via `--vars` overrides any existing values in the environment, meaning `VAR=1 ain template.ain --vars VAR=2` will result in VAR having the value `2`.

Ain looks for any .env file in the folder where it's run for any default variable values. You can pass the path to a custom .env file via the `-e` flag.

Environment variables are replaced before executables and can be used as input to the executable. Example `$(cat ${ENV}/token.json)`.

Ain uses [envparse](https://github.com/hashicorp/go-envparse) for parsing .env files.

# Executables
An executable expression (example `$(command arg1 arg2)`) will be replaced by running the command with arguments and replacing the expression with the commands output (STDOUT). For example `$(echo 1)` will be replaced by `1`.

A real world example is getting JWT tokens from a separate script and share that across templates:
```
[Headers]
Authorization: Bearer $(bash -c "./get-login.sh | jq -r '.token'")
```

If shell features such as pipes are needed this can be done via a command string (e g [bash -c](https://man7.org/linux/man-pages/man1/bash.1.html#OPTIONS)) in bash. Note that quoting is needed if the argument contains whitespace as in the example above. See [quoting](#quoting).

The first word is an command on your $PATH and the rest are arguments to that command.

See [escaping](#escaping) for arguments containing closing-parentheses `)`.

Executables are replaced after environment-variables.

# Fatals
Ain has two types of errors: fatals and errors. Errors are things internal to ain (it's not your fault) such as not finding the backend-binary.

Fatals are errors in the template (it's your fault). Fatals include the template file name where the fatal occurred, the line-number and a small context of the template:
```
$ ain templates/example.ain
Fatal error in file: templates/example.ain
Cannot find value for variable PORT on line 2:
1   [Host]
2 > http://localhost:${PORT}
3
```

Fatals can be hard to understand if [environment variables](#environment-variables) or [executables](#executables) are replaced in the template. If the line with the fatal contains any replaced value a separate expanded context is printed. It contains up to three lines with the resulting replacement and the row number into the original template:
```
$ TIMEOUT=-1 ain templates/example.ain 
Fatal error in file: templates/example.ain
Timeout interval must be greater than 0 on line 10:
9   [Config]
10 > Timeout=${TIMEOUT}
11   
Expanded context:
10 > Timeout=-1
```

# Quoting
There are four places where quoting might be necessary: arguments to executables, backend options, invoking the $VISUAL or $EDITOR command and when passing template-names via a pipe. All for the same reasons as bash: a word is an argument to something and a whitespace is the delimiter to the next argument. If whitespace should be retained it must be quoted.

The canonical example of when quoting is needed is doing more complex things involving pipes. E g `$(sh -c 'find . | fzf -m | xargs echo')`.

Escaping is kept simple, you can use `\'` or `\"` respectively to insert a literal quote inside a quoted string of the same type. You can avoid this by selecting the other quote character (e g 'I need a " inside this string') when possible.

# Escaping
TL;DR: To escape a comment `#` precede it with a backtick: `` `#``.

These symbols have special meaning to ain: 
```
Symbol -> meaning
#      -> comment
${     -> environment variable
$(     -> executable
```

If you need these symbols literally in your output, escape with a backtick:
```
Symbol -> output
`#     -> #
`${    -> ${
`$(    -> $(
```

If you need a literal backtick just before a symbol, you escape the escaping with a slash:
```
\`#
\`${
\`$(
```

If you need a literal `}` in an environment variable you escape it with a backtick:
```
Template    -> Environment variable
${VA`}RZ}   -> VA}RZ
```

If you need a literal `)` in an executable, either escape it with a backtick or enclose it in quotes.
These two examples are equivalent and inserts the string Hi:
```
$(node -e console.log('Hi'`))
$(node -e 'console.log("Hi")')
```

If you need a literal backtick right before closing the envvar or executable you escape the backtick with a slash:
```
$(echo \`)
${VAR\`}
```

Since environment variables are only replaced once, `${` doesn't need escaping when returned from an environment variable. E g `VAR='${GOAT}'`, `${GOAT}` is passed literally to the output. Same for executables, any returned value containing `${` does not need escaping. E g `$(echo $(yo )`, `$(yo ` is passed literally to the output.

Pound sign (#) needs escaping if a comment was not intended when returned from both environment variables and executables.

A section header (one of the eight listed under [supported sections](#supported-sections)) needs escaping if it's the only text a separate line. It is escaped with a backtick. Example:
```
[Body]
I'm part of the
`[Body]
and included in the output.
```

If you need a literal backtick followed by a valid section heading you escape that backtick with a slash. Example:
```
[Body]
This text is outputted as
\`[Body]
backtick [Body].
```

# URL-encoding
Both the path and the query-section of an url is scanned and any invalid characters are [URL-encoded](https://en.wikipedia.org/wiki/Percent-encoding)  while already legal encodings (format `%<hex><hex>` and `+` for the query string) are kept as is.

This means that you can mix url-encoded text, half encoded text or unencoded text and ain will convert everything into a properly url-encoded URL.

Example:
```
[Host]
https://localhost:8080/api/finance/ca$h

[Query]
account=full of ca%24h   # This is already url-encoded (%24 = $)
```

Will result in the URL:
```
https://localhost:8080/api/finance/ca%24h?account=full+of+ca%24h
```

The only caveats is that ain cannot know if a plus sign (+) is an encoded space or an literal plus sign. In this case ain assumes a space and leave the plus sign as is.

Second ain cannot know if you meant the literal percent sign followed by two hex characters %<hex><hex> instead of an encoded percent character. In this case ain assumes an escaped sequence and leaves the %<hex><hex> as is.

In both cases you need to manually escape the plus (%2B) and percent sign (%25) in the url.

# Sharing is caring
Ain can print out the command instead of running it via the `-p` flag. This enables you to inspect how the curl, wget or httpie API call would look like:
```
ain -p base.ain create-blog-post.ain > share-me.sh
```

The output can then be shared (or for example run over an ssh connection).

Piping it into bash is equivalent to running the command without `-p`.
```
ain -p base.ain create-blog-post.ain | bash
```

Any content within the [[Body]](#Body) section when passing the flag `-p` will be written to a file in the current working directory where ain is invoked. The file is not removed after ain completes. See [[Body]](#body) for details.

# Handling line endings
Ain uses line-feed (\n) when printing it's output. If you're on windows and storing ain:s result to a file, this
may cause trouble. Instead of trying to guess what line ending we're on (WSL, docker, cygwin etc makes this a wild goose chase), you'll have to manually convert them if the receiving program complains.

Instructions here: https://stackoverflow.com/a/19914445/1574968

# Troubleshooting
If the templates are valid but the actual backend call fails, passing the `-p` flag will show you the command ain tries to run. Invoking this yourself in a terminal might give you more clues to what's wrong.

# Ain in a bigger context
But wait! There's more!

With ain being terminal friendly there are few neat tricks in the [wiki](https://github.com/jonaslu/ain/wiki)

# Contributing
I'd love if you want to get your hands dirty and improve ain!

If you look closely there are almost* no tests. There's even a [commit](9e114a3) wiping all tests that once was. Why is a good question. WTF is also a valid response.

It's an experiment you see, I've blogged about [atomic literate commits](https://www.iamjonas.me/2021/01/literate-atomic-commits.html) paired with a thing called a [test plan](https://www.iamjonas.me/2021/04/the-test-plan.html). This means you make the commit solve one problem, write in plain english what problem is and how the commit solves it and how you verified that it works. All of that in the commit messages. For TL;DR; do a `git log` and see for yourself.

I'll ask you to do the same and we'll experiment together. See it as a opportunity to try on something new.

\* Except for where it does make sense to have a unit-test: to exercise a well known algo and prove it's correct as done in utils_test.go. Doing this by hand would be hard, timeconsuming and error prone.
