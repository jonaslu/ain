# What problem does ain solve?
It's a front-end for curl and httpie. The insomnia / paw equivalent
for the terminal. Scripting and pipes are encouraged.

# GIF
<Screen recording of cat:ing a .env, global.ain, then a local.ain
then running it. Then fzf with ain and some more fancy stuff>.

# The longer story
I often need copy data coming from some script into the insomnia or paw,
then copy-paste the result back into a file for further processing
on the command line. It's a manual pipe. This is not ok.

Most projects have at least one backend with at minimum 10-20 endpoints
returning data. I can't recall all of the endpoints and their expected formats.

Bash makes having a common base hard (including scripts in scripts is painful)
I wind up with 10-20 copy-pasted scripts, one for each endpoint.
One change ripples through all 10-20 of them.

The backend takes some form of Authorization: header, most common is a JWT<link>,
 that needs fetching and refreshing. Quoting and bash just won't let me
pass headers into curl or httpie. I've tried. Thus I copy paste this too
into each script.

Ain solves this by combining files with increased specificity. It thrives in
a pipe and sharing of scripts. It makes the actual http(s) call via curl
or httipe returning the result for further processing.

# Nomenclature
Templates: Files containing what, how and where to make the http-call.
Sections: Headings in a template file.
Environment variables: Adding variables to a template-file.
Subshells: Adding in results of a script into a template-file.
Backends: The thing that makes the http-call (curl or httipe).
Fatals: When the parsing of template files go wrong (it's your fault).

# Templates
Ain works with files. These files are called templates. Here's an example:
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
-sS
# Comments are ignored
```
The template file can be named anything so naming the contents is
entirely up to you.

The template uses the old DOS [ini]-format where the sections (the things in square brackets)
are meta-data and the text below a section is the data associated with that meta-data.

Each of the sections are described in details below.

Ain reads templates via the command line or via a pipe
(if it's connected to one). Sections combine over all the files. This enables
you to have a base-file with global settings, and then one or more files
that specify things that differ between the files.

An example would be setting the Headers, [Backend] and [BackendOptions] in
a base-file and then specifying the specific [Host], [Method] and [Body] in a
specific file.

If several and the same section appears in both files, the latter section
overrides the former section (but not [Host], [Headers] and [BackendOpts] which
append the latter to the former - more on that below).

Anything after a pound sign (#) is a comment and is ignored.

# Invocation
ain is used as such:
`ain [options] <template-files...>[!]`

It accepts options and one or more template files (that override or
accumulate if several files have the same sections).

Suffixing the file-name with an exclamation-mark (!) makes ain open the file in
your $EDITOR (or vim if not set) so you can edit the file one-shot
before executing.

This editing is not stored to the file and used only for this invocation.

If ain is connected to a pipe it will read file-names off that pipe.
This enables you to use find<link> and a selector such as fzf<link> to keep
track of the template-files: `find . -name *.ain | fzf -m | ain`

# Sections

## [Host]
Contains the URL to call. This header appends lines from file to file.

This neat little feature allows you to specify a base-urlin one file
(e g base.ain) as such:
`http://localhost:3000`

And in the next file specify the endpoint (e g login.ain): `/api/auth/login`

And you could have query parameters in yet another file (pwd-login.ain):
```
?user=leviathan
&password=dearMother
```

This way base-urls can be re-used.

## [Headers]
Just what it sounds like, what headers to include in the call.

Headers append line-by-line too so you can share common headers (e g
Authorization: <JWT> and Content-Type: application/json)

## [Method]
What http-method to use in the call. If omitted defalts to whatever the
backend defaults to when not specified (GET in both curl and httpie).

## [Body]
If the http-call needs a body (POST, PATCH) the content of this section
is passed as a file to the backend.

The reason for always using a file and not compressed into a mangled
string is that formatting may be important (as it is in yaml).

This means that json can be pretty-printed so you can eyeball it easily.
This section overwrites any previous [Body] sections found in earlier files.

The temp-file is removed after the call unless you pass the -l
flag. It places the tempfile in the $TMPFILE directory (usually
/tmp on your box). You can override this in your shell-environment
if you'd like them elsewhere.

## [Config]
This section contains any config mean for ain (any backend-specific config is
passed via the [BackendOptions] section).

Currently the only option supported is Timeout=<timeout in seconds>

The timeout is enforced during the whole execution of ain
(both running subshells and the backend-call).

If omitted defaults to infinity (no timeout).

## [Backend]
The backend is the command used to run the actual http-call.

Valid options are curl<link> or httpie<link>.

## [BackendOptions]
Any options meant for the backends. These are appended straight to the
backend-command invocation (curl or httpie).

# Environment variables
Ain supports variables in the templates via environment variables.
Anything inside ${} is replaced with the value found in the environment.

Ain also reads any .env files in the folder from where it's run. You
can pass a custom .env file via the -e flag.

This enables you to specify things that vary across calls either
permanently in the .env file or one-shot via the command-line:
`PORT=5000 ain global.ain local.ain`

# Subshells
Ain supports running scripts and injecting the result of the script
into a template.

Anything inside $() is replaced with the result from running that command
and capturing it's output.

This enables you to put things such as getting JWT tokens into a separate
script and share that across templates.

More complex scripting can be done in-line with the xargs `bash -c` hack<ink>:
```
[Headers]
Authorization: Bearer $(bash -c "./get-login.sh | jq -r '.token')
```

# Sharing is caring
Ain can print out the command instead of running
it via the -p flag. This enables you to
inspect how the curl or httpie call would look like
or to share the command:
```
ain -p base.ain create-blog-post.ain > share-me.sh
```

Piping it into bash is equivalent to running the command
without -p.
```
ain -p base.ain create-blog-post.ain | bash
```

# Ain in a bigger context - make this a wiki
But wait! There's more!

With ain being terminal friendly there are few neat tricks in the wiki you
might want to check out: <link>

## Turn the finding .ain files into an alias

## Global alias
Instead of having an .rc file specifying global settings you can simply an
alias with your own global settings:
`alias ain='ain .global.ain'`

## Diffing results
By using temp-file redirection you can now compare results from different
invocation. This is great for comparing your local result from an API with
some reference:
`meld <(ENV=test ain service.ain) <(ENV=prod ain service.ain)`
