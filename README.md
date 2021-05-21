# What problem does ain solve?
It's an API client for the terminal. Scripts and pipes welcome!

# GIF
<Screen recording of cat:ing a .env, global.ain, then a local.ain
then running it. Then fzf with ain and some more fancy stuff>.

# Quick start
Ain uses sections in square brackets to specify what API to call.

Start by putting things common to a service in a file (let's call
it base.ain):

```
$> cat base.ain
[Host]
http://localhost:8080

[Headers]
Content-Type: application/json

[Backend]
curl
```

Then add another file for specific url:
```
$> cat create-blog-post.ain
[Host]
/api/blog/create

[Body]
{
  "title": "Million dollar idea",
  "text": "A dating service. With music."
}
```

Run ain to combine them into a single call and print the result:
```
$> ain base.ain create-blog-post.ain
{
  "status": "ok"
}
```

# Installation
`go get -u github.com/jonaslu/ain`

# Pre-requisites
Go (version 1.11 or higher).

You need curl and or httpie installed on your machine and available on your $PATH (command.Exec needs to find the binary).

# Important concepts
* Templates: Files containing what, how and where to make the http-call. By convention has the file-ending `.ain`.
* Sections: Headings in a template file.
* Environment variables: Enables variables in a template file.
* Subshells: Enables using the results of a script in a template file.
* Backends: The thing that makes the http-call (curl or httipe).
* Fatals: Error in parsing the template files (it's your fault).

# Templates
Ain reads what to call from template-files. Here's a full example:
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
  "title": "New million-dollar idea",
  "content": "A social networking platform. With music."
}

[Config]
Timeout=10

[Backend]
curl

[BackendOptions]
-sS # Comments are ignored.
# This too.
```
The template files can be named anything but some unique ending-convention such as .ain is encouraged. Then you can search for templates across the entire filesystem and then pipe them into ain.

Ain understands seven [Sections] (the things in square brackets). Each of the sections are described in details [below](#supported-sections).

Ain reads template file-names from the command line or via a pipe. Sections either combine or overwrite across all the given template files. This enables you to have one common base template file with global settings, and then one or more template files that specify things that differ between API-calls (such as URL-path or query parameters).

An example would be setting the [Headers], [Backend] and [BackendOptions] in a base template file and then specifying the specific [Host], [Method] and [Body] in several template files, one for each API-endpoint.

Anything after a pound sign (#) is a comment and will be ignored.

# Running ain
`ain [options] <template-files...>[!]`

Ain accepts options and one or more template files. Adding an exclamation-mark (!) at the end of the template file name makes ain open the file in your $EDITOR (or vim if not set) so you can edit the file. The edit is not stored back into the template file and used only for this API-call.

If ain is connected to a pipe it will try to read template file names off that pipe. This enables you to use [find](https://man7.org/linux/man-pages/man1/find.1.html) and a selector such as [fzf](https://github.com/junegunn/fzf)> to keep track of the template-files: `find . -name *.ain | fzf -m | ain`

Template file names specified on the command line are read before any names from a pipe. This means that in the `echo create-blog-post.ain | ain base.ain` is the same as `ain base.ain create-blog-post.ain`.

# Supported sections


## [Host]
Contains the URL to call. This section appends the lines one template file to the next.

This neat little feature allows you to specify a base-url in one file
(e g base.ain) as such:
`http://localhost:3000`

And in the next template file specify the endpoint (e g login.ain): `/api/auth/login`

And you could have query parameters in yet another template file (e g pwd-login.ain):
```
?user=leviathan
&password=dearMother
```

This way base-urls can be re-used.

This section is mandatory.

## [Headers]
What headers to include in the call.

The [Headers] section appends across template files so you can share common headers (e g Authorization: <JWT> and Content-Type: application/json)

## [Method]
What http-method to use in the call (GET, POST, PATCH). If omitted defalts to whatever the backend defaults to when not specified (GET in both curl and httpie).

The [Method] section is overridden by latter template files.

## [Body]
If the API call needs a body (POST, PATCH) the content of this section is passed as a file to the backend.

The reason for always using a file is that formatting may be important (e g yaml). In the template file this section can be pretty-printed (e g json).

The file passed to the backendn is removed after the call unless you pass the `-l` flag. Ain places the file in the $TMPFILE directory (usually `/tmp` on your box). You can override this in your shell if you'd like them elsewhere.

The [Body] sections is overridden by latter template files.

## [Config]
This section contains config for ain (any backend-specific config is
passed via the [BackendOptions] section).

Currently the only option supported is Timeout=<timeout in seconds>

The timeout is enforced during the whole execution of ain
(both running subshells and the backend-call).

If omitted defaults to no timeout.

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
Ain supports variables in the templates via environment variables.
Anything inside ${} is replaced with the value found in the environment.

Ain also reads any .env files in the folder from where it's run. You
can pass a custom .env file via the `-e` flag.

This enables you to specify things that vary across calls either
permanently in the .env file or one-shot via the command-line. Example:
`PORT=5000 ain base.ain create-blog-post.ain`

# Subshells
Anything inside a $() is replaced with the result from running that shell-command
and capturing it's output (STDIN).

This enables you to put things such as getting JWT tokens into a separate
script and share that across templates.

More complex scripting can be done in-line with the xargs `bash -c` hack<ink>. Example:
```
[Headers]
Authorization: Bearer $(bash -c "./get-login.sh | jq -r '.token'")
```
# Fatals
Ain separates fatals from errors. Errors are things internal to ain (it's not your fault) such as not finding the backend-binary.

Fatals are errors in the template (it's your fault). Ain will try to parse as much of the templates as possible aggregating fatals before reporting back to you. Fatals include the template file name where the fatal occurred, line-number and a small context of the template:
```
$ ain templates/example.ain
Error in file: templates/example.ain
Fatal error Cannot find value for variable PORT on line 2:
1   [Host]
2 > http://localhost:${PORT}
3
```

# Sharing is caring
Ain can print out the command instead of running it via the `-p` flag. This enables you to
inspect how the curl or httpie call would look like or to share the command:
```
ain -p base.ain create-blog-post.ain > share-me.sh
```

Piping it into bash is equivalent to running the command without `-p`.
```
ain -p base.ain create-blog-post.ain | bash
```
