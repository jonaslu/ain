# What problem does ain solve?
It's an API client for the terminal. Scripts and pipes welcome!
![Show and tell](/assets/show-and-tell.gif?raw=true)

# Quick start
Ain uses sections in square brackets to specify what API to call.

Start by putting things common to a service in a file (let's call it base.ain):

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

See all options: `ain -h`

# Installation
`go get -u github.com/jonaslu/ain`

# Pre-requisites
Go (version 1.13 or higher).

You need curl and or httpie installed on your machine and available on your $PATH (command.Exec needs to find the binary).

# Important concepts
* Templates: Files containing what, how and where to make the API call. By convention has the file-ending `.ain`.
* Sections: Headings in a template file.
* Environment variables: Enables variables in a template file.
* Executables: Enables using the results of another command in a template file.
* Backends: The thing that makes the API call (curl or httipe).
* Fatals: Error in parsing the template files (it's your fault).

# Templates
Ain reads sections from template-files. Here's a full example:
```
[Host]
http://localhost:${PORT}/api/blog/post

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
The template files can be named anything but some unique ending-convention such as .ain is recommended so you can [fing](https://man7.org/linux/man-pages/man1/find.1.html) them easily.

Ain understands seven [Sections] (the things in square brackets). Each of the sections are described in details [below](#supported-sections).

Sections either combine or overwrite across all the template files given to ain.

Anything after a pound sign (#) is a comment and will be ignored.

# Running ain
`ain [options] <template-files...>[!]`

Ain accepts one or more template-files as a mandatory parameter. As sections combine or overwrite where it makes sense you can better organize API-calls into hierarchical structures with increasing specificity. An example would be setting the [Headers], [Backend] and [BackendOptions] in a base template file and then specifying the specific [Host], [Method] and [Body] in several template files, one for each API-endpoint. You can even use an `alias` for things you will always set.

Adding an exclamation-mark (!) at the end of the template file name makes ain open the file in your `$EDITOR` (or vim if not set) so you can edit the file. The edit is not stored back into the template file and used only this invocation.

If ain is connected to a pipe it will try to read template file names off that pipe. This enables you to use [find](https://man7.org/linux/man-pages/man1/find.1.html) and a selector such as [fzf](https://github.com/junegunn/fzf) to keep track of the template-files: `find . -name *.ain | fzf -m | ain`

Template file names specified on the command line are read before any names from a pipe. This means that `echo create-blog-post.ain | ain base.ain` is the same as `ain base.ain create-blog-post.ain`.

# Supported sections

## [Host]
Contains the URL to the API. This section appends the lines from one template file to the next. This neat little feature allows you to specify a base-url in one file (e g base.ain) as such: `http://localhost:3000` and in the next template file specify the endpoint (e g login.ain): `/api/auth/login`

You could have query parameters in yet another template file (e g user-leviathan-login.ain):
```
?user=leviathan
&password=dearMother
```

The [Host] section is mandatory and appends across template files.

## [Headers]
Headers to include in the API call.

The [Headers] section appends across template files so you can share common headers (e g Authorization: <JWT> and Content-Type: application/json)

## [Method]
What http-method to use in the API call (e g GET, POST, PATCH). If omitted defalts to whatever the backend defaults to when not specified (GET in both curl and httpie).

The [Method] section is overridden by latter template files.

## [Body]
If the API call needs a body (POST, PATCH) the content of this section is passed as a file to the backend because formatting may be important (e g yaml). In the template file this section can be pretty-printed for easier eye-balling (e g json).

The file passed to the backend is removed after the API call unless you pass the `-l` flag. Ain places the file in the $TMPFILE directory (usually `/tmp` on your box). You can override this in your shell by explicitly setting `$TMPFILE` if you'd like them elsewhere.

The [Body] sections is overridden by latter template files.

## [Config]
This section contains config for ain (any backend-specific config is passed via the [BackendOptions] section).

Currently the only option supported is `Timeout=<timeout in seconds>`

The timeout is enforced during the whole execution of ain (both running executables and the actual API call). If omitted defaults to no timeout. Note that this is the only section where executables have
no effect, since the timeout needs to be known before the executables are invoked.

The [Cnnfig] sections is overridden by latter template files.

## [Backend]
The [Backend] specifies what command should be used to run the actual API call.

Valid options are [curl](https://curl.se/) or [httpie](https://httpie.io/).

The [Backend] section is mandatory and is overridden by latter template files.

## [BackendOptions]
Any options meant for the backends. These are appended straight to the
backend-command invocation (curl or httpie).

The [BackendOptions] section appends across template files.

# Environment variables
Anything inside `${}` in a template is replaced with the value found in the environment.

Ain also reads any .env files in the folder from where it's run. You can pass a custom .env file via the `-e` flag.

This enables you to specify things that vary across API calls either permanently in the .env file or one-shot via the command-line. Example:
`PORT=5000 ain base.ain create-blog-post.ain`

Environment-variables are expanded first and can be used with any executable. Example `$(cat ${ENV}/token.json)`.

Ain uses [envparse](https://github.com/hashicorp/go-envparse) for, well, parsing environment variables, so anything it can do, so can you!

# Executables
Anything inside a `$()` is replaced with the result from running that command and capturing it's output (STDIN). The command can return multiple rows which will be inserted as separate rows in the template (e g returning two headers).

An example is getting JWT tokens into a separate script and share that across templates.

More complex scripting can be done in-line with the xargs `bash -c` (hack)[https://en.wikipedia.org/wiki/Xargs#Shell_trick:_any_number]. Example:
```
[Headers]
Authorization: Bearer $(bash -c "./get-login.sh | jq -r '.token'")

Ain expects the first word in an executable to be on your $PATH and the rest to be arguments (hence the need for quotes to bash -c as this is passed as one argument).

Executables are captured and replaced in the template after any environment-variables so if the script returns an environment-variable name it won't be expanded into any value.
```

# Fatals
Ain has two types of errors: fatals and errors. Errors are things internal to ain (it's not your fault) such as not finding the backend-binary.

Fatals are errors in the template (it's your fault). Ain will try to parse as much of the templates as possible aggregating fatals before reporting back to you. Fatals include the template file name where the fatal occurred, line-number and a small context of the template:
```
$ ain templates/example.ain
Fatal error in file: templates/example.ain
Cannot find value for variable PORT on line 2:
1   [Host]
2 > http://localhost:${PORT}
3
```

# Sharing is caring
Ain can print out the command instead of running it via the `-p` flag. This enables you to inspect how the curl or httpie API call would look like or share the command:
```
ain -p base.ain create-blog-post.ain > share-me.sh
```

Piping it into bash is equivalent to running the command without `-p`.
```
ain -p base.ain create-blog-post.ain | bash
```

A note on line-endings. Ain uses line-feed (\n) when printing it's output. If you're on windows and storing ain:s result to a file, this
may cause trouble. Instead of trying to guess what line ending we're on (WSL, docker, cygwin etc makes this a wild goose chase), you'll have to manually convert them if the receiving progrogram complains.

Instructions here: https://stackoverflow.com/a/19914445/1574968

# Ain in a bigger context
But wait! There's more!

With ain being terminal friendly there are few neat tricks in the [wiki](https://github.com/jonaslu/ain/wiki)
