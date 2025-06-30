" ~/.vim/syntax/ain.vim
if exists("b:current_syntax")
  finish
endif

" Headings
syntax match ainHeading /^\s*\[\(config\|host\|query\|headers\|method\|body\|backend\|backendoptions\)\]\s*\ze\(\s*#\|\s*$\)\c/
highlight link ainHeading Keyword

" Escapes
syntax match ainEscape /\\`/
syntax match ainEnvvarEscape /`\${/
syntax match ainEscape /`\$(/
syntax match ainEscape /`#/
highlight link ainEscape Normal
highlight link ainEnvvarEscape Normal

" Envvars: ${VAR}
syntax region ainEnvvar start=+\${+ end=+}+ contains=ainEnvvarEndEscape
syntax match ainEnvvarEndEscape /`}/ contained
highlight link ainEnvvar Keyword
highlight link ainEnvvarEndEscape Keyword

" Executables: $(command)
syntax region ainExec start=+\$(+ end=+)+ contains=ainEscape,ainEnvvarEscape,ainEnvvar,ainExecEscape,ainSQ,ainDQ
syntax match ainExecEscape /`)/ contained
highlight link ainExec Identifier
highlight link ainExecEscape Identifier
highlight link ainEnvvarEscape Identifier
highlight link ainEscape Identifier

" Single-quoted strings inside executables
syntax region ainSQ start=+'+ end=+'+ contains=ainEscape,ainEnvvarEscape,ainEnvvar,ainSQEscape contained
syntax match ainSQEscape /\\'/ contained
highlight link ainSQ String
highlight link ainSQEscape String
highlight link ainEnvvarEscape String
highlight link ainEscape String

" Double-quoted strings inside executables
syntax region ainDQ start=+"+ end=+"+ contains=ainEscape,ainEnvvarEscape,ainEnvvar,ainDQEscape contained
syntax match ainDQEscape /\\"/ contained
highlight link ainDQ String
highlight link ainDQEscape String
highlight link ainEnvvarEscape String
highlight link ainEscape String

" Comments # comment
syntax match ainComment /#.*/
highlight link ainComment Comment

let b:current_syntax = "ain"
