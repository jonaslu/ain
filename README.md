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
- [Longer start](#longer-start)
- [Important concepts](#important-concepts)
- [Templates](#templates)
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
- [Environment variables](#environment-variables)
- [Executables](#executables)
- [Fatals](#fatals)
- [URL-encoding](#url-encoding)
- [Sharing is caring](#sharing-is-caring)
- [Troubleshooting](#troubleshooting)
- [Ain in a bigger context](#ain-in-a-bigger-context)
- [Contributing](#contributing)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Pre-requisites
You need [curl](https://curl.se/), [wget](https://www.gnu.org/software/wget/) or [httpie](https://httpie.io/) installed and available on your `$PATH`. The easiest way to test this is to run `ain -b`. This will generate a basic starter template listing what backends you have available on your system in the [[Backend]](#backend) section. It will select one and leave the others commented out.

You can also check manually what backends you have installed by opening up a shell and type `curl`, `wget` or `http` (add the suffix .exe to those commands if you're on windows). If there's any output from the command itself you're good to go.

On linux or mac one of the three above is very likely to be installed on your box already. The others are available in your package manager or [homebrew](https://brew.sh).

If you're on windows curl.exe is already installed if it's windows 10 build 17063 or higher. Otherwise you can get the binaries via [scoop](https://scoop.sh), [chocolatey](https://chocolatey.org/) or download them yourself. Ain uses curl.exe and cannot use the curl cmd-let powershell builtin.

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
Ain comes with a built in basic template that you can use as a starting point. Ain also checks what backends (that's [curl](https://curl.se/), [wget](https://www.gnu.org/software/wget/) or [httpie](https://httpie.io/)) are available on your system and inserts them into the [[Backend]](#backend) section of the generated template. One will be selected and the rest commented out so the template is runnable directly.

Run:
```
ain -b basic_template.ain
```

The command above will output a starter-template to the file `basic_template.ain`.
The basic template contains a common scenario of calling GET on localhost
with the `Content-Type: application/json`.

Run the generated template by specifying a `PORT` environment variable:
```
PORT=8080 ain basic_template.ain
```

# Longer start
Ain uses sections in square brackets to specify how to call an API.

Start by putting things common to all APIs for a service in a file (let's call it base.ain):

```
$> cat base.ain
[Host]
http://localhost:8080

[Headers]
Content-Type: application/json

[Backend]
curl

[BackendOptions]
-sS
```

Then add another file for a specific URL:
```
$> cat create-blog-post.ain
[Host]
/api/blog/create

[Method]
POST

[Body]
{
  "title": "Million dollar idea",
  "text": "A dating service. With music."
}
```

Run ain to combine them into a single API call and print the result:
```
$> ain base.ain create-blog-post.ain
{
  "status": "ok"
}
```

See the help for all options ain supports: `ain -h`

# Important concepts
* Templates: Files containing what, how and where to make the API call. By convention has the file-ending `.ain`.
* Sections: Headings in a template file.
* Environment variables: Enables variables in a template file.
* Executables: Enables using the results of another command in a template file.
* Backends: The thing that makes the API call ([curl](https://curl.se/), [wget](https://www.gnu.org/software/wget/) or [httpie](https://httpie.io/)).
* Fatals: Error in parsing the template files (it's your fault).

# Templates
Ain reads sections from template-files. Here's a full example:
```
[Host]
http://localhost:${PORT}/api/blog/post

[Query]
id=2e79870c-6504-4ac6-a2b7-01da7a6532f1

[Headers]
Authorization: Bearer $(./get-jwt-token.sh)
Content-Type: application/json

[Method]
POST

[Body]
{
  "title": "Reaping death",
  "content": "There is a place beyond the dreamworlds past the womb of night."
}

[Config]
Timeout=10

[Backend]
curl

[BackendOptions]
-sS # Comments are ignored.
# This too.
```
The template files can be named anything but some unique ending-convention such as .ain is recommended so you can [find](https://man7.org/linux/man-pages/man1/find.1.html) them easily.

Ain understands eight [Sections] (the things in square brackets). Each of the sections are described in details [below](#supported-sections).

Sections either combine or overwrite across all the template files given to ain.

Anything after a pound sign (#) is a comment and will be ignored.

# Running ain
`ain [options] <template-files...>[!]`

Ain accepts one or more template-file(s) as a mandatory parameter. As sections combine or overwrite where it makes sense you can better organize API-calls into hierarchical structures with increasing specificity.

An example would be setting the [[Headers]](#Headers), [[Backend]](#backend) and [[BackendOptions]](#BackendOptions) in a base template file and then specifying the specific [[Host]](#Host), [[Method]](#Method) and [[Body]](#Body) in several template files, one for each API-endpoint. You can even use an [alias](https://www.gnu.org/software/bash/manual/html_node/Aliases.html) for things you will always set.

Adding an exclamation-mark (!) at the end of the template file name makes ain open the file in your `$VISUAL` or `$EDITOR` editor or falls back to vim in that order so you can edit the template file. Any changes are not stored back into the template file and used only this invocation.

Example:
```
ain templates/get-blog-post.ain!
```

Ain waits for the editor command to exit. Any terminal editor such as vim, emacs, nano etc will be fine. If your editor of choice forks (e g [vscode](https://code.visualstudio.com/) does by default) check if there's a flag stopping it from forking. For example to stop vscode from forking use the `--wait` [flag](https://code.visualstudio.com/docs/editor/command-line#_core-cli-options):

```
export EDITOR="code --wait"
```

If ain is connected to a pipe it will try to read template file names off that pipe. This enables you to use [find](https://man7.org/linux/man-pages/man1/find.1.html) and a selector such as [fzf](https://github.com/junegunn/fzf) to keep track of the template-files:
```
$> find . -name *.ain | fzf -m | ain
```

Template file names specified on the command line are read before any names from a pipe. This means that `echo create-blog-post.ain | ain base.ain` is the same as `ain base.ain create-blog-post.ain`.

Ain functions as bash when it comes to file names: if they contain white-space the name should be quoted.

When making the call ain mimics how data is returned by the backend. After printing any internal errors of it's own, ain echoes back output from the backend: first the standard error (stderr) and then the standard out (stdout). It then returns the exit code from the backend command as it's own unless there are error specific to ain in which it returns status 1.

# Supported sections
Sections are case-insensitive and whitespace ignored but by convention uses CamelCase and are left indented. A section cannot be defined twice in a file. A section ends where the next begins or the file ends.

In the unlikely event that the contents of a section must contain the exact same text as a valid section (one of the eight below) on one line you can escape that text with a `\` and it will be passed as text.

E g:
```
[Body]
All of this will be passed as the body!
\[Body]
Including the text above.
```

## [Host]
Contains the URL to the API. This section appends the lines from one template file to the next. This neat little feature allows you to specify a base-url in one file (e g `base.ain`) as such: `http://localhost:3000` and in the next template file specify the endpoint (e g `login.ain`): `/api/auth/login`.

It's recommended that you use the [[Query]](#Query) section below for query-parameters as it handles joining with delimiters and trimming whitespace. You  can however put raw query-paramters in the [Host] section too.

Any query-parameters added in the [[Query]](#Query) section are appended last to the URL. The whole URL is properly [url-encoded](#url-encoding) before passed to the backend. The [Host] section shall combine to one and only one valid URL. Multiple URLs is not supported.

Ain performs no validation on the url (as backends differ on what a valid url looks like). If your call does not go through use `ain -p` as mentioned in [troubleshooting](#troubleshooting) and input that directly into the backend to see what it thinks it means.

The [Host] section is mandatory and appends across template files.

## [Query]
All lines in the [Query] section is appended to the URL after the complete URL has been assembled. This means that you can specify query-parameters that apply to many endpoints in one file instead of having to include the same parameter in all endpoints.

An example is if an `API_KEY=<secret>` query-parameter applies to several endpoints. You can define this query-parameter in a base-file and simply have the specific endpoint URL and possible extra query-parameters in their own file.

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

To avoid the common bash-ism error of having spaces around the equals sign, the whitespace in a query key / value is only significant within the string.

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
What http-method to use in the API call (e g GET, POST, PATCH). If omitted the backend default is used (GET in both curl, wget and httpie).

Example:
```
[Method]
POST
```

The [Method] section is overridden by latter template files.

## [Body]
If the API call needs a body (as in the POST or PATCH http methods) the content of this section is passed as a file to the backend with the formatting retained from the [Body] section. Ain uses files to pass the [Body] contents because white-space may be important (e g yaml) and this section tends to be long.

The file passed to the backend is removed after the API call unless you pass the `-l` (as in leave) flag. Ain places the file in the $TMPDIR directory (usually `/tmp` on your box). You can override this in your shell by explicitly setting `$TMPDIR` if you'd like them elsewhere.

Passing print command `-p` (as in print) flag will cause ain to write out the file named ain-body<random-digits> in the directory where ain is invoked (`cwd`) and leave the file after completion. The `-p` flag is for [[sharing]](#sharing-is-caring) and for [[troubleshooting]](#troubleshooting). Leaving the body file makes the resulting printed command shareable and runnable.

The [Body] section removes any trailing whitespace and keeps empty newlines between the first and last non-empty line.

Example:
```
[Body]

{
  "some": "json",  # ain removes comments

  "more": "jayson"
}

```

Is passed as this in the tmp-file:
```
{

  "some": "json",

  "more": "jayson"
}
```

The [Body] section is overridden by latter template files.

## [Config]
This section contains config for ain. All config parameters are case-insensitive and any whitespace is ignored. Parameters for backends themselves are passed via the [[BackendOptions]](#BackendOptions) section.

Full config example:
```
[Config]
Timeout=3
QueryDelim=;
```

The [Config] sections is overridden by latter template files.

### Timeout
Config format: `Timeout=<timeout in seconds>`

The timeout is enforced during the whole execution of ain (both running executables and the actual API call). If omitted defaults to no timeout. This is the only section where [executables](#executables) cannot be used, since the timeout needs to be known before the executables are invoked.

### Query delimiter
Config format: `QueryDelim=<text>`

This is the delimiter used when concatenating the lines under the [[Query]](#Query) section to form the query-string of an URL. It can be any text that does not contain a space including the empty string.

It defaults to (`&`).

## [Backend]
The [Backend] specifies what command should be used to run the actual API call.

Valid options are [curl](https://curl.se/), [wget](https://www.gnu.org/software/wget/) or [httpie](https://httpie.io/).

Example:
```
[Backend]
curl
```

The [Backend] section is mandatory and is overridden by latter template files.

## [BackendOptions]
Backend specific options that are passed on to the backend command invocation.

Example:
```
[Backend]
curl

[BackendOptions]
-sS   # Makes curl disable it's progress bar in a pipe
```

The [BackendOptions] section appends across template files.

# Environment variables
Anything inside `${}` in a template is replaced with the value found in the environment.

Ain also reads any .env files in the folder from where it's run. You can pass a custom .env file via the `-e` flag. Only new variables are set. Any already existing env-variable is not modified.

This enables you to specify things that vary across API calls either permanently in the .env file or one-shot via the command-line. Example:
`PORT=5000 ain base.ain create-blog-post.ain`

Environment-variables are expanded first and can be used with any executable. Example `$(cat ${ENV}/token.json)`.

Ain uses [envparse](https://github.com/hashicorp/go-envparse) for parsing environment variables.

# Executables
An executable expression (i e `$(command arg1 arg2)`) will be replaced by running the command with any arguments and replacing the expression with the output (STDOUT). For example `$(echo 1)` will be replaced by `1` when processing the template.

An more real world example is getting JWT tokens from a separate script and share that across templates:
```
[Headers]
Authorization: Bearer $(bash -c "./get-login.sh | jq -r '.token'")
```

If shell features such as pipes are needed this can be done via a command string (e g [bash -c](https://man7.org/linux/man-pages/man1/bash.1.html#OPTIONS) in bash.

If parentheses are needed as arguments they must be within quotes (e g $(node -e 'console.log("Hi")')") to not end the executable expression.

Ain expects the first word in an executable to be on your $PATH and the rest to be arguments (hence the need for quotes to bash -c as this is passed as one argument).

Executables are captured and replaced in the template after any environment-variables are expanded. This means that anything the executable returns is inserted directly even if it's a valid environment variable name.

# Fatals
Ain has two types of errors: fatals and errors. Errors are things internal to ain (it's not your fault) such as not finding the backend-binary.

Fatals are errors in the template (it's your fault). Ain will try to parse as much of the templates as possible aggregating fatals before reporting back to you. Fatals include the template file name where the fatal occurred, the line-number and a small context of the template:
```
$ ain templates/example.ain
Fatal error in file: templates/example.ain
Cannot find value for variable PORT on line 2:
1   [Host]
2 > http://localhost:${PORT}
3
```

Fatals can be hard to understand if [environment variables](#environment-variables) or [executables](#executables) substitute for values in the template. If the line with the fatal contains any substituted value a separate expanded context is printed. It contains up to three lines with the resulting substitution and a row number into the original template:
```
$ TIMEOUT=-1 go run cmd/ain/main.go templates/example.ain 
Fatal error in file: templates/example.ain
Timeout interval must be greater than 0 on line 10:
9   [Config]
10 > Timeout=${TIMEOUT}
11   
Expanded context:
10 > Timeout=-1
```

# Quoting
Quoting in bash is hard and therefore ain tries avoid it. There are four places where it might be necessary: Arguments to executables, backend options, invoking the $VISUAL or $EDITOR command and when passing template-names via a pipe into ain. All for the same reasons as bash: a word is an argument to something and a whitespace is the delimiter to the next argument. If whitespace is part of the argument it must be explicit.

The canonical example of when quoting is needed is doing more complex things involving pipes. E g `$(sh -c 'find . | fzf -m | xargs echo')`.

Quoting is kept simple, you can use ' or ". There is only one escape-sequence (`\'` and `\"` respectively) to insert a quote inside a quoted string of the same type. You can avoid when possible by selecting the other quote character (e g 'I need a " inside this string').

# URL-encoding
[URL-encoding](https://en.wikipedia.org/wiki/Percent-encoding) is something ain tries hard to take care of for you. Both the path and the query-section of an url is scanned and any non-valid charaters are encoded while already legal encodings (format `%<hex><hex>` and `+` for the query string) are kept as is.

This means that you can mix url-encoded text, half encoded text or unencoded text and ain will convert them all into a properly url-encoded URL.

Example:
```
[Host]
https://localhost:8080/api/finance/ca$h

[Query]
account=full of ca%24h
```

Will result in the URL:
```
https://localhost:8080/api/finance/ca%24h?account=full+of+ca%24h
```

The only caveats is that ain cannot know if a plus sign (+) is an encoded space or if the actual plus sign was meant. In this case ain leaves the plus sign as is. Also it cannot know if you actually meant %<hex><hex> instead of an encoded character. In both cases you need to yourself manually escape the plus (%2B) and percent sign (%25).

# Sharing is caring
Ain can print out the command instead of running it via the `-p` flag. This enables you to inspect how the curl, wget or httpie API call would look like or share the command:
```
ain -p base.ain create-blog-post.ain > share-me.sh
```

Piping it into bash is equivalent to running the command without `-p`.
```
ain -p base.ain create-blog-post.ain | bash
```

Any content within the [[Body]](#Body) section when passing the flag `-p` will be written to a file in the current working directory where ain is invoked. The file is not removed after ain completes. See [[Body]](#body) for details.

# Handling line endings
A note on line-endings. Ain uses line-feed (\n) when printing it's output. If you're on windows and storing ain:s result to a file, this
may cause trouble. Instead of trying to guess what line ending we're on (WSL, docker, cygwin etc makes this a wild goose chase), you'll have to manually convert them if the receiving program complains.

Instructions here: https://stackoverflow.com/a/19914445/1574968

# Troubleshooting
If the templates are valid but the actual backend call fails, passing the `-p` flag will show you  the command ain tries to run. Invoking this yourself in a terminal might give you more clues to what's wrong.

# Ain in a bigger context
But wait! There's more!

With ain being terminal friendly there are few neat tricks in the [wiki](https://github.com/jonaslu/ain/wiki)

# Contributing
I'd love if you want to get your hands dirty and improve ain!

If you look closely there are almost* no tests. There's even a [commit](9e114a3) wiping all tests that once was. Why is a good question. WTF is also a valid response.

It's an experiment you see, I've blogged about [atomic literate commits](https://www.iamjonas.me/2021/01/literate-atomic-commits.html) paired with a thing called a [test plan](https://www.iamjonas.me/2021/04/the-test-plan.html). This means you make the commit solve one problem, write in plain english what problem is and how the commit solves it and how you verified that it works. All of that in the commit messages. For TL;DR; do a `git log` and see for yourself.

I'll ask you to do the same and we'll experiment together. See it as a opportunity to try on something new.

\* Except for where it does make sense to have a unit-test: to exercise a well known algo  and prove it's correct as done in utils_test.go. Doing this by hand would be hard, timeconsuming and error prone.
